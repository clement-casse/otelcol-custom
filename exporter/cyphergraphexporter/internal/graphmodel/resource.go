package graphmodel

import (
	"context"
	_ "embed"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.opentelemetry.io/collector/pdata/pcommon"
)

//go:embed mergeResources.cypher
var cypherQueryNewResource string

func (e *Encoder) MergeResource(ctx context.Context, tx neo4j.ManagedTransaction, r *pcommon.Resource) error {
	// Forge a list of map[string]any to be sent as query parameter of neo4j query
	// each entry in rParams is a distinct resource that will be MERGEd with existing
	// one in the graph. Merging occurs on resources sharing both the same "label" and "id" keys.
	// Each resource is a map[string]any expecting the following keys:
	// - "label" designates the kind of resource e.g. a K8S pod, a VM, a container
	// - "id" designates its unique identifier in the system
	// - "props" is another map[string]string with additional parameter to be added to the resource as additional context.
	// - "contained_in" links to the node that contains this resource by identifying it by both its label and id
	rParams := make([]map[string]any, 0, len(e.resourceByAttrKeys))
	for attrKey, re := range e.resourceByAttrKeys {
		// Check if key attribute is present in pdata's resource
		if attrValue, ok := r.Attributes().Get(string(attrKey)); ok {
			props := make(map[string]any, len(re.AdditionalPropKeys))
			for _, ap := range re.AdditionalPropKeys {
				if value, ok := r.Attributes().Get(string(ap)); ok {
					props[string(ap)] = value.AsString()
				}
			}
			var containedIn []map[string]any
			// Check the existence of the resources that contain this resource from the containment hierarchy and the resource attributes.
			if dependantResourceTypes, ok := e.containmentOrder[re.ResourceType]; ok {
				containedIn = make([]map[string]any, 0, len(dependantResourceTypes))
				for _, dependantResourceType := range dependantResourceTypes {
					// find which resource attribute is used as key to merge the containing resource
					if dependantResAttrKey, ok := e.attributesByLabels[dependantResourceType]; ok {
						// get the value associated of this key in the resource
						if dependantResourceID, ok := r.Attributes().Get(string(dependantResAttrKey)); ok {
							containedIn = append(containedIn, map[string]any{
								"label": dependantResourceType,
								"id":    dependantResourceID.AsString(),
							})
							break
						}
						e.logger.Debug("expecting the resource to have attribute '%v' to identify the containing resource '%s', attribute was not found, discarding the containing resource", dependantResAttrKey, dependantResourceType)
					} else {
						e.logger.Warnf("'%s' in an unknown resource type and its key attribute cannot be found", dependantResourceType)
					}
				}
			} else {
				containedIn = []map[string]any{}
				e.logger.Debugf("there is no known resource hierarchy for '%s'", re.ResourceType)
			}
			rParams = append(rParams, map[string]any{
				"label":        re.ResourceType,
				"id":           attrValue.AsString(),
				"props":        props,
				"contained_in": containedIn,
			})
		} else {
			e.logger.Debugf("field %s not found in telemetry's resource, skipping", string(attrKey))
		}
	}
	// Execute a cypher query on the list of resources that UNWINDs the full list.
	result, err := tx.Run(ctx, cypherQueryNewResource, map[string]any{"resources": rParams})
	if err != nil {
		return err
	}
	_, err = result.Consume(ctx)
	return err
}
