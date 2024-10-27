package cyphergraphexporter // import "github.com/clement-casse/otelcol-custom/exporter/cyphergraphexporter"

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/exporter/exporterhelper"

	"github.com/clement-casse/otelcol-custom/exporter/cyphergraphexporter/internal/metadata"
)

// NewFactory creates a factory for the CypherGraph exporter.
func NewFactory() exporter.Factory {
	return exporter.NewFactory(
		metadata.Type,
		createDefaultConfig,
		exporter.WithTraces(createTracesExporter, metadata.TracesStability))
}

func createDefaultConfig() component.Config {
	return &Config{
		DatabaseURI:     defaultDatabaseURI,
		UserAgent:       defaultUserAgent,
		ResourceMappers: defaultResourcesMappers,
	}
}

func createTracesExporter(ctx context.Context, set exporter.Settings, cCfg component.Config) (exporter.Traces, error) {
	cfg := cCfg.(*Config)
	exp, err := newTracesExporter(cfg, set)
	if err != nil {
		return nil, err
	}
	return exporterhelper.NewTraces(ctx, set, cfg,
		exp.tracesPusher,
		exporterhelper.WithCapabilities(consumer.Capabilities{MutatesData: false}),
		exporterhelper.WithStart(exp.Start),
		exporterhelper.WithShutdown(exp.Shutdown),
	)
}
