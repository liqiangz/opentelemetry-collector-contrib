module github.com/open-telemetry/opentelemetry-collector-contrib/examples/demo/server

go 1.17

require (
	go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.26.0
	go.opentelemetry.io/otel v1.1.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.24.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.24.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.1.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.1.0
	go.opentelemetry.io/otel/metric v0.24.0
	go.opentelemetry.io/otel/sdk v1.1.0
	go.opentelemetry.io/otel/sdk/metric v0.24.0
	go.opentelemetry.io/otel/trace v1.1.0
	google.golang.org/grpc v1.52.3
)

require (
	github.com/cenkalti/backoff/v4 v4.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/felixge/httpsnoop v1.0.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.16.0 // indirect
	go.opentelemetry.io/otel/internal/metric v0.24.0 // indirect
	go.opentelemetry.io/otel/sdk/export/metric v0.24.0 // indirect
	go.opentelemetry.io/proto/otlp v0.9.0 // indirect
	golang.org/x/net v0.4.0 // indirect
	golang.org/x/sys v0.3.0 // indirect
	golang.org/x/text v0.5.0 // indirect
	google.golang.org/genproto v0.0.0-20221118155620-16455021b5e6 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)
