module github.com/desyang-hub/go-rpc

go 1.21

require (
	google.golang.org/grpc v1.65.0
	google.golang.org/protobuf v1.34.2
	go.opentelemetry.io/otel v1.29.0
	go.opentelemetry.io/otel/sdk v1.29.0
	go.opentelemetry.io/otel/trace v1.29.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.29.0
	go.opentelemetry.io/otel/exporters/prometheus v0.49.0
	github.com/prometheus/client_golang v1.20.0
	github.com/rs/zerolog v1.33.0
	github.com/hashicorp/consul/api v1.30.0
	go.etcd.io/etcd/client/v3 v3.5.15
	github.com/spf13/cobra v1.8.1
	github.com/spf13/viper v1.19.0
	github.com/stretchr/testify v1.9.0
	go.uber.org/mock v0.4.0
)
