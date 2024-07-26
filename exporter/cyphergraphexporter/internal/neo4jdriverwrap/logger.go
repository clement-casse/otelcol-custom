package neo4jdriverwrap

import (
	"fmt"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j/log"
	"go.uber.org/zap"
)

// NewLogger takes a *zap.SugaredLogger to create a logger compliant with the interface Logger from
// "github.com/neo4j/neo4j-go-driver/v5/neo4j/log".
func NewLogger(zsl *zap.SugaredLogger) log.Logger {
	return &logger{zsl: zsl}
}

// logger wraps the zap logger to comply neo4j Log interface "github.com/neo4j/neo4j-go-driver/v5/neo4j/log"
type logger struct {
	zsl *zap.SugaredLogger
}

func (l logger) Error(name, id string, err error) {
	l.zsl.Named(name).Errorf("id=%s %s", id, err)
}

func (l logger) Warnf(name, id, msg string, args ...any) {
	l.zsl.Named(name).Warnf(fmt.Sprintf("id=%s %s", id, msg), args...)
}

func (l logger) Infof(name, id, msg string, args ...any) {
	l.zsl.Named(name).Infof(fmt.Sprintf("id=%s %s", id, msg), args...)
}

func (l logger) Debugf(name, id, msg string, args ...any) {
	l.zsl.Named(name).Debugf(fmt.Sprintf("id=%s %s", id, msg), args...)
}
