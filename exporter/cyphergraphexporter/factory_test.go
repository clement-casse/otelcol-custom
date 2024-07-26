package cyphergraphexporter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/exporter/exportertest"
)

// withDefaultConfig create a new default configuration
// and applies provided functions to it.
func withDefaultConfig(fns ...func(*Config)) *Config {
	cfg := createDefaultConfig().(*Config)
	for _, fn := range fns {
		fn(cfg)
	}
	return cfg
}

func TestCreateDefaultConfig(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	assert.NotNil(t, cfg)
	assert.NoError(t, componenttest.CheckConfigStruct(cfg))
}

func TestFactory_CreateMetricsExporter_Fail(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	params := exportertest.NewNopSettings()
	_, err := factory.CreateMetricsExporter(context.Background(), params, cfg)
	require.Error(t, err)
}

func TestFactory_CreateLogsExporter_Fail(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	params := exportertest.NewNopSettings()
	_, err := factory.CreateLogsExporter(context.Background(), params, cfg)
	require.Error(t, err)
}

func TestFactory_CreateTraceExporter_Fail(t *testing.T) {
	factory := NewFactory()
	cfg := withDefaultConfig(func(c *Config) {
		c.DatabaseURI = ""
	})
	params := exportertest.NewNopSettings()
	_, err := factory.CreateTracesExporter(context.Background(), params, cfg)
	require.Error(t, err)
}

func TestFactory_CreateTraceExporter(t *testing.T) {
	factory := NewFactory()
	cfg := factory.CreateDefaultConfig()
	params := exportertest.NewNopSettings()
	exporter, err := factory.CreateTracesExporter(context.Background(), params, cfg)
	require.NoError(t, err)
	require.NotNil(t, exporter)

	require.NoError(t, exporter.Shutdown(context.TODO()))
}
