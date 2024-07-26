package cyphergraphexporter // import "github.com/clement-casse/playground/otelcol-custom/exporter/cyphergraphexporter"

import (
	"context"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/exporter"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"

	"github.com/clement-casse/playground/otelcol-custom/exporter/cyphergraphexporter/internal/graphmodel"
	"github.com/clement-casse/playground/otelcol-custom/exporter/cyphergraphexporter/internal/neo4jdriverwrap"
)

type cyphergraphTraceExporter struct {
	driver  neo4j.DriverWithContext
	encoder *graphmodel.Encoder

	logger *zap.Logger
}

func newTracesExporter(cfg *Config, set exporter.Settings) (cte *cyphergraphTraceExporter, err error) {
	cte = &cyphergraphTraceExporter{logger: set.Logger}
	if err = cfg.Validate(); err != nil {
		return
	}
	var neo4jAuth neo4j.AuthToken
	if cfg.Username != "" {
		neo4jAuth = neo4j.BasicAuth(cfg.Username, string(cfg.Password), "")
	} else if t := string(cfg.BearerToken); t != "" {
		neo4jAuth = neo4j.BearerAuth(t)
	} else if t := string(cfg.KerberosTicket); t != "" {
		neo4jAuth = neo4j.KerberosAuth(t)
	} else {
		neo4jAuth = neo4j.NoAuth()
	}
	cte.driver, err = neo4j.NewDriverWithContext(
		cfg.DatabaseURI,
		neo4jAuth,
		neo4jdriverwrap.WithLogger(set.Logger),
		neo4jdriverwrap.WithUserAgent(cfg.UserAgent),
		neo4jdriverwrap.WithBackOffConfig(&cfg.BackOffConfig),
	)
	if err != nil {
		return
	}
	labelFromAttr := make(map[attribute.Key]graphmodel.ResourceEncoder, len(cfg.ResourceMappers))
	for label, matcher := range cfg.ResourceMappers {
		labelFromAttr[attribute.Key(matcher.IdentifiedByKey)] = graphmodel.ResourceEncoder{
			ResourceType:       label,
			AdditionalPropKeys: make([]attribute.Key, len(matcher.OtherProperties)),
		}
		for i, prop := range matcher.OtherProperties {
			labelFromAttr[attribute.Key(matcher.IdentifiedByKey)].AdditionalPropKeys[i] = attribute.Key(prop)
		}
	}
	cte.encoder = graphmodel.NewEncoder(labelFromAttr, defaultContainmentOrder, set.Logger)
	return
}

func (e *cyphergraphTraceExporter) Start(ctx context.Context, _ component.Host) error {
	err := e.driver.VerifyConnectivity(ctx)
	if err != nil {
		return err
	}
	// TODO think about initializing database indices here.
	return nil
}

func (e *cyphergraphTraceExporter) Shutdown(ctx context.Context) error {
	if e.driver == nil {
		return nil
	}
	return e.driver.Close(ctx)
}

func (e *cyphergraphTraceExporter) tracesPusher(ctx context.Context, td ptrace.Traces) error {
	start := time.Now()
	session := e.driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	rSpans := td.ResourceSpans()
	for i := 0; i < rSpans.Len(); i++ {
		e.logger.Sugar().Infof("span %d, Span: %+v", i, rSpans.At(i))
		// _, err := neo4j.ExecuteWrite(ctx, session, func(tx neo4j.ManagedTransaction) (any, error) {
		// 	resource := rSpans.At(i).Resource()
		// 	if err := graphmodel.MergeResource(ctx, tx, &resource); err != nil {
		// 		return nil, err
		// 	}

		// 	return nil, nil
		// })
		// if err != nil {
		// 	e.logger.Error("resources merge error", zap.Error(err))
		// }
	}

	duration := time.Since(start)
	e.logger.Debug("traces inserted", zap.String("duration", duration.String()))
	return nil
}
