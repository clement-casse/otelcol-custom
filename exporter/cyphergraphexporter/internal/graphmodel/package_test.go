package graphmodel

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/goleak"
)

var (
	neo4jContainerImage    = "neo4j:5-bullseye"
	memgraphContainerImage = "memgraph/memgraph:latest"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

// The following are utility functions used throughout the various tests present in this package.

// resourceFrom returns a [go.opentelemetry.io/collector/pdata/pcommon.Resource] with the provided Key Value pairs.
func resourceFrom(attr []attribute.KeyValue) pcommon.Resource {
	res := pcommon.NewResource()
	for _, kv := range attr {
		switch kv.Value.Type() {
		case attribute.BOOL:
			res.Attributes().PutBool(string(kv.Key), kv.Value.AsBool())
		case attribute.INT64:
			res.Attributes().PutInt(string(kv.Key), kv.Value.AsInt64())
		case attribute.FLOAT64:
			res.Attributes().PutDouble(string(kv.Key), kv.Value.AsFloat64())
		case attribute.STRING:
			res.Attributes().PutStr(string(kv.Key), kv.Value.AsString())
		case attribute.BOOLSLICE, attribute.INT64SLICE, attribute.FLOAT64SLICE, attribute.STRINGSLICE:
			panic("slice attributes are not managed by the resourceFrom function")
		default:
			panic("Type not handled by resourceFrom function: " + kv.Value.Type().String())
		}
	}
	return res
}

// InitCypherDBTestContainer is a helper function that starts a compatible database in a TestContainer to execute Cypher
// queries on it for testing.
//
//	ctx := context.Background()
//	driver, closeFunc, err := InitCypherDBTestContainer(ctx, engine)
//	if err != nil {
//		panic(err)
//	}
//	defer closeFunc(ctx)
//	err = driver.VerifyConnectivity(ctx)
//	if err != nil {
//		panic(err)
//	}
func InitCypherDBTestContainer(ctx context.Context, engine string) (neo4j.DriverWithContext, func(context.Context), error) {
	var req testcontainers.ContainerRequest
	switch engine {
	case "neo4j":
		req = testcontainers.ContainerRequest{
			Image:        neo4jContainerImage,
			ExposedPorts: []string{"7687/tcp"},
			Env:          map[string]string{"NEO4J_AUTH": "none"},
			WaitingFor:   wait.ForLog("Started."),
		}
	case "memgraph":
		req = testcontainers.ContainerRequest{
			Image:        memgraphContainerImage,
			ExposedPorts: []string{"7687/tcp"},
			Cmd:          []string{"--bolt-port=7687", "--telemetry-enabled=false"},
			WaitingFor:   wait.ForLog("You are running Memgraph"),
		}
	default:
		return nil, nil, fmt.Errorf("engine '%s' not supported, only acceptable values are ['neo4j', 'memgraph']", engine)
	}

	dbContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, nil, err
	}
	mappedPort, err := dbContainer.MappedPort(ctx, "7687")
	if err != nil {
		return nil, nil, err
	}
	hostIP, err := dbContainer.Host(ctx)
	if err != nil {
		return nil, nil, err
	}
	targetDBPort := fmt.Sprintf("bolt://%s:%s", hostIP, mappedPort)
	driver, err := neo4j.NewDriverWithContext(targetDBPort, neo4j.NoAuth())
	if err != nil {
		return nil, nil, err
	}
	closeFunc := func(ctx context.Context) {
		if errDefer := driver.Close(ctx); errDefer != nil {
			panic(errDefer)
		}
		if errDefer := dbContainer.Terminate(ctx); errDefer != nil {
			panic(errDefer)
		}
	}
	time.Sleep(500 * time.Millisecond) // a shame, but the wait.ForLog(`...`) seems a bit clunky and the initialization occasionnaly fails

	return driver, closeFunc, nil
}
