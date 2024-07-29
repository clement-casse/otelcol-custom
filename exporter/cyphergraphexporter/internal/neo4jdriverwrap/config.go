package neo4jdriverwrap

import (
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/config"
	"go.opentelemetry.io/collector/config/configretry"
	"go.uber.org/zap"
)

// WithUserAgent
func WithUserAgent(ua string) func(*config.Config) {
	return func(cfg *config.Config) {
		cfg.UserAgent = ua
	}
}

// WithLogger
func WithLogger(zl *zap.Logger) func(*config.Config) {
	l := NewLogger(zl.Sugar())
	return func(cfg *config.Config) {
		cfg.Log = l
	}
}

// WithBackOffConfig
func WithBackOffConfig(boc *configretry.BackOffConfig) func(*config.Config) {
	return func(cfg *config.Config) {
		if boc == nil {
			return
		}
		if !boc.Enabled {
			return
		}
		cfg.MaxTransactionRetryTime = boc.MaxInterval
	}
}
