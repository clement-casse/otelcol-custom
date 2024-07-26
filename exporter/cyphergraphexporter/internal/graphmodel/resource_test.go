package graphmodel

import (
	"context"
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.uber.org/zap"
)

const databaseEngine = "memgraph"

var (
	emptyResource  = pcommon.NewResource()
	simpleResource = resourceFrom([]attribute.KeyValue{
		semconv.ServiceName("name-of-the-service"),
	})
	k8sResource = resourceFrom([]attribute.KeyValue{
		semconv.ServiceName("my-deployment"),
		semconv.K8SClusterName("my-cluster"),
		semconv.K8SNamespaceName("my-namespace"),
		semconv.K8SDeploymentName("my-deployment"),
		semconv.K8SReplicaSetName("my-deployment-66cf4d99b5"),
		semconv.K8SPodName("my-deployment-66cf4d99b5-kpqg"),
		semconv.K8SPodUID("7293ca81-d35e-459d-b15a-a8197fbc03e4"),
	})
)

func TestMergeResource(t *testing.T) {
	ctx := context.Background()
	driver, closeFunc, err := InitCypherDBTestContainer(ctx, databaseEngine)
	assert.NoError(t, err)
	defer closeFunc(ctx)
	logger, _ := zap.NewDevelopment()

	encoder := NewEncoder(map[attribute.Key]ResourceEncoder{
		semconv.K8SPodNameKey:        {"k8s.pod", []attribute.Key{semconv.K8SPodUIDKey}},
		semconv.K8SDeploymentNameKey: {"k8s.deployment", []attribute.Key{semconv.K8SDeploymentUIDKey}},
		semconv.K8SClusterNameKey:    {"k8s.cluster", []attribute.Key{semconv.K8SClusterUIDKey}},
	}, map[string][]string{
		"k8s.pod":        {"k8s.deployment"},
		"k8s.deployment": {"k8s.cluster"},
		"k8s.cluster":    {},
	}, logger)

	for _, tt := range []struct {
		name         string
		resource     *pcommon.Resource
		cypherQuery  string
		expectResult [][]any
	}{
		{
			name:         "an empty resource creates nothing",
			resource:     &emptyResource,
			cypherQuery:  `MATCH (r:Resource) RETURN r`,
			expectResult: [][]any{},
		}, {
			name:        "create a simple resource with no containment relations",
			resource:    &simpleResource,
			cypherQuery: `MATCH (r:Resource) RETURN r.type, r.id`,
			expectResult: [][]any{
				{string(semconv.ServiceNameKey), "name-of-the-service"},
			},
		}, {
			name:        "check K8S containment",
			resource:    &k8sResource,
			cypherQuery: `MATCH (r1:Resource)-[:IS_CONTAINED]->(r2:Resource) RETURN r1.type, r1.id, r2.type, r2.id`,
			expectResult: [][]any{
				{"k8s.pod", "my-deployment-66cf4d99b5-kpqg", "k8s.deployment", "my-deployment"},
				{"k8s.deployment", "my-deployment", "k8s.cluster", "my-cluster"},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			session := driver.NewSession(ctx, neo4j.SessionConfig{})
			_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
				return nil, encoder.MergeResource(ctx, tx, tt.resource)
			})
			assert.NoError(t, err)

			_, err = session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
				result, err2 := tx.Run(ctx, tt.cypherQuery, map[string]any{})
				assert.NoError(t, err2)
				records, err2 := result.Collect(ctx)
				assert.NoError(t, err2)

				for _, rec := range records {
					assert.Contains(t, tt.expectResult, rec.Values)
				}
				return nil, nil
			})
			assert.NoError(t, err)
		})
	}

}
