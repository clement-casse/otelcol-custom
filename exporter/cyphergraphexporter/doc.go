//go:generate go run go.opentelemetry.io/collector/cmd/mdatagen metadata.yaml

// Package cyphergraphexporter implements an OpenTelemetry Collector exporter that sends trace data to
// a graph database using the Cypher language (Neo4j, Memgraph, ...)
package cyphergraphexporter // import "github.com/clement-casse/otelcol-custom/exporter/cyphergraphexporter"
