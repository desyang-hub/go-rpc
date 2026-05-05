# Service Registration Guide

This guide covers configuring service registration with go-rpc.

## Configuration

### Consul Backend

```yaml
discovery:
  backend: consul
  consul:
    address: "consul-server:8500"
    scheme: "https"
    datacenter: "dc1"
    token: "your-consul-token"
    service:
      name: "my-service"
      id: "my-service-1"
      port: 50051
      tags: ["v1", "production"]
      meta:
        version: "1.0.0"
        zone: "us-east-1a"
      health_check:
        tcp:
          interval: 10s
          timeout: 5s
        http:
          path: "/health"
          interval: 15s
```

### etcd Backend

```yaml
discovery:
  backend: etcd
  etcd:
    endpoints:
      - "etcd-1:2379"
      - "etcd-2:2379"
      - "etcd-3:2379"
    key_prefix: "/services"
    tls:
      ca_file: "/path/to/ca.crt"
      cert_file: "/path/to/client.crt"
      key_file: "/path/to/client.key"
    service:
      name: "my-service"
      port: 50051
      weight: 100
```

## Go Configuration

```go
registryConfig := consul.Config{
    Address:  "consul-server:8500",
    Scheme:   "https",
    DataSource: "dc1",
    Service: consul.ServiceConfig{
        Name: "my-service",
        Port: 50051,
        Tags: []string{"v1", "production"},
    },
}

registry := consul.NewRegistry(registryConfig)
```

## Service ID Generation

For multi-instance deployments, generate unique service IDs:

```go
func generateServiceID(name string) string {
    hostname, _ := os.Hostname()
    pid := os.Getpid()
    return fmt.Sprintf("%s-%s-%d", name, hostname, pid)
}
```

## Health Check

go-rpc provides built-in health check endpoints:

```go
import "github.com/desyang-hub/go-rpc/pkg/health"

healthMiddleware := health.NewMiddleware(
    health.WithTCPCheck(":50052"),
    health.WithHTTPHealthEndpoint("/health"),
)
```

### Health Check Paths

| Endpoint | Purpose |
|----------|---------|
| `/health` | Overall health status |
| `/ready` | Readiness check (notifies Consul/etcd) |
| `/livez` | Liveness probe |

## Next Steps

- [Load Balancing Setup](load-balancing.md) — Client-side load distribution
- [Observability](observability.md) — Monitoring and metrics
