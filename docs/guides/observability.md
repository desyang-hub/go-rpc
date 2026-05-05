# Observability

This guide covers configuring observability with OpenTelemetry, Prometheus, and structured logging.

## OpenTelemetry Tracing

### Configuration

```yaml
# config.yaml
tracing:
  enabled: true
  exporter: otlp
  endpoint: "otel-collector:4317"
  sample_rate: 1.0  # 100% sampling
  service_name: "my-service"
```

### Go Configuration

```go
import "github.com/desyang-hub/go-rpc/pkg/opentelemetry"

tracer, err := opentelemetry.NewTracer(
    "my-service",
    opentelemetry.WithEndpoint("otel-collector:4317"),
    opentelemetry.WithSamplingRate(1.0),
)
if err != nil {
    log.Fatal(err)
}

srv := server.NewServer().
    Tracer(tracer).
    Build()
```

### Trace Context Propagation

go-rpc automatically propagates trace context through gRPC metadata:

```
grpc-trace-bin: <serialized trace context>
grpc-timeout: 10m
```

## Prometheus Metrics

### Configuration

```yaml
# config.yaml
metrics:
  enabled: true
  addr: ":9090"
  path: "/metrics"
  namespace: "rpc"
```

### Collected Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `rpc_client_request_duration_seconds` | Histogram | Request latency |
| `rpc_client_request_total` | Counter | Total requests |
| `rpc_client_request_failed_total` | Counter | Failed requests |
| `rpc_server_request_duration_seconds` | Histogram | Server processing time |
| `rpc_server_active_connections` | Gauge | Active connections |

### Go Configuration

```go
import "github.com/desyang-hub/go-rpc/pkg/prometheus"

metrics := prometheus.NewMetrics("my-service")

// Expose metrics endpoint
go metrics.Serve(":9090")

srv := server.NewServer().
    Metrics(metrics).
    Build()
```

## Structured Logging

### Configuration

```yaml
# config.yaml
logging:
  level: info
  format: json
  output: stdout
  caller: false
```

### Log Levels

| Level | Usage |
|-------|-------|
| `debug` | Detailed diagnostic information |
| `info` | General operational messages |
| `warn` | Non-critical issues |
| `error` | Error conditions |
| `fatal` | Fatal errors (with exit) |

### Example JSON Log Output

```json
{
  "level": "info",
  "ts": "2025-01-15T10:30:00Z",
  "msg": "request completed",
  "service": "hello-service",
  "rpc": "/hello.HelloService/Hello",
  "method": "Hello",
  "duration_ms": 12.5,
  "status": "OK",
  "client_ip": "192.168.1.100"
}
```

### Go Configuration

```go
import "github.com/desyang-hub/go-rpc/pkg/logging"

logger := logging.New(
    logging.WithLevel("info"),
    logging.WithFormat("json"),
)

srv := server.NewServer().
    Logger(logger).
    Build()
```

## Integration Example

```go
// Initialize observability components
tracer := opentelemetry.New("my-service")
metrics := prometheus.New("my-service")
logger := logging.New(logging.WithLevel("info"))

// Wire into server
srv := server.NewServer().
    Address(":50051").
    Tracer(tracer).
    Metrics(metrics).
    Logger(logger).
    AddMiddleware(middleware.Logging()).
    AddMiddleware(middleware.Tracing(tracer)).
    Build()
```

## Grafana Dashboard

Common dashboard panels:

| Panel | Visualization | Metric |
|-------|---------------|--------|
| Request Rate | Time series | `rpc_client_request_total` |
| Error Rate | Gauge | `rpc_client_request_failed_total` |
| Latency P99 | Histogram | `rpc_client_request_duration_seconds` |
| Active Connections | Gauge | `rpc_server_active_connections` |

## Next Steps

- [Authentication Guide](authentication.md) — Secure your services
- [Kubernetes Deployment](../deployment/kubernetes.md) — Deploy to K8s
