ARG go_version=1.23
ARG otelcol_name=otelcol
ARG otelcol_builder_version=0.129.0

FROM powerman/dockerize:0.19.0 AS dockerize

# Use the Go image to run the OpenTelemetry Collector Builder Binary that will generate
# and compile Go code to create the collector binary.
FROM golang:${go_version} AS otelcolbuilder

ARG otelcol_name
ARG otelcol_builder_version

RUN --mount=type=cache,target=/go/pkg/mod \
    go install github.com/mikefarah/yq/v4@v4.45.1; \
    go install go.opentelemetry.io/collector/cmd/builder@v${otelcol_builder_version};

WORKDIR /usr/src/app

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    yq -i '.dist.output_path = "/usr/src/gen" | \
    .dist.name ="'${otelcol_name}'"' ./builder-config.yaml; \
    CGO_ENABLED=0 GOOS=linux GOWORK=off \
    builder --config=./builder-config.yaml;

# Use a Distroless-base image to run collector with minimal environment
# (the distroless/static-debian12 one is not sufficient and the collector cannot be run).
FROM gcr.io/distroless/base-debian12

ARG otelcol_name

WORKDIR /

COPY --from=dockerize --chown=nonroot:nonroot \
    /usr/local/bin/dockerize /usr/local/bin/

COPY --from=otelcolbuilder --chown=nonroot:nonroot \
    /usr/src/gen/${otelcol_name} /otelcol

USER nonroot:nonroot

ENTRYPOINT ["/otelcol"]

LABEL org.opencontainers.image.source=https://github.com/clement-casse/otelcol-custom
LABEL org.opencontainers.image.licenses=MIT
LABEL org.opencontainers.image.description="An OpenTelemetry Collector built with custom modules for research purpose."
