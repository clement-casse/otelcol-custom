package graphmodel

import (
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

// Encoder defines the configuration required to encode traces in a hierarchical property graph
type Encoder struct {
	// resourceByAttrKeys associates the OpenTelemetry Attributes Keys that uniquely identify resources of the given label.
	resourceByAttrKeys map[attribute.Key]ResourceEncoder

	// containmentOrder represents the graph of resource containment as an adjacency list:
	// e.g.
	//    containmentOrder := map[string][]string{
	//        "k8s.pod":                 {"k8s.node"},
	//        "k8s.node":                {"k8s.cluster", "cloud.availability.zone"},
	//        "cloud.availability.zone": {"cloud.region"},
	//        "cloud.region":            {},
	//        "k8s.cluster":             {},
	//    }
	containmentOrder map[string][]string

	// attributesByLabels does a reverse indexation of Node labels to the corresponding attribute.Key
	attributesByLabels map[string]attribute.Key

	logger *zap.SugaredLogger
}

// ResourceEncoder is a struct that identifies a type of resource and all the attributes found in telemetry
// data that will be used to create a vertex in graph encoding procedure.
type ResourceEncoder struct {
	ResourceType       string
	AdditionalPropKeys []attribute.Key
}

// NewEncoder creates a graph encoder with the provided parameters
func NewEncoder(labelFromAttr map[attribute.Key]ResourceEncoder, containmentOrder map[string][]string, logger *zap.Logger) *Encoder {
	abl := make(map[string]attribute.Key, len(labelFromAttr))
	for key, re := range labelFromAttr {
		abl[re.ResourceType] = key
	}
	return &Encoder{
		resourceByAttrKeys: labelFromAttr,
		containmentOrder:   containmentOrder,
		attributesByLabels: abl,
		logger:             logger.Sugar().Named("graphencoder"),
	}
}
