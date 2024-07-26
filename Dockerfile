FROM golang:1.21 AS otelcolbuilder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
COPY exporter/cyphergraphexporter/go.mod exporter/cyphergraphexporter/go.sum ./exporter/cyphergraphexporter/

RUN --mount=type=cache,target=$GOPKG/pkg \
    go mod download && go mod verify

COPY . .
RUN --mount=type=cache,target=$GOPKG/pkg \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux \
    go run github.com/mikefarah/yq/v4 -i '.dist.output_path = "/usr/src/app/dist"' ./builder-config.yaml; \
    go run go.opentelemetry.io/collector/cmd/builder --config=./builder-config.yaml

ENTRYPOINT ["/usr/src/app/dist/otelcol-custom-dev"]
