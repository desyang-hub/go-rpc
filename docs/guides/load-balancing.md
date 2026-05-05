# Load Balancing Setup

This guide covers configuring load balancing strategies for your go-rpc services.

## Configuration

### Round Robin

```yaml
# config.yaml
loadbalancer:
  strategy: round_robin
```

### Weighted Round Robin

```yaml
loadbalancer:
  strategy: weighted_round_robin
  weights:
    "192.168.1.10:50051": 100
    "192.168.1.11:50051": 50
    "192.168.1.12:50051": 25
```

### Least Connections

```yaml
loadbalancer:
  strategy: least_connections
  min_weight: 0  # Ignore endpoints below this weight
```

### Consistent Hashing

```yaml
loadbalancer:
  strategy: consistent_hashing
  hash_key: "user_id"  # Field name in request for session affinity
```

## Go Configuration

```go
lbConfig := loadbalancer.Config{
    Strategy: loadbalancer.WeightedRoundRobin,
    Weights: map[string]int{
        "192.168.1.10:50051": 100,
        "192.168.1.11:50051": 50,
    },
    HealthCheckInterval: 10 * time.Second,
    Jitter: true,
}

balancer := loadbalancer.New(lbConfig)
```

## Client Configuration

```go
cl := client.NewClient().
    Address("192.168.1.10:50051,192.168.1.11:50051").
    LoadBalancer(balancer).
    Build()
```

## Weighted Balancing with Discovery

When using service discovery with weighted balancing:

1. Server registers with weight in metadata
2. Client reads weight from discovery
3. Requests distributed proportionally

```go
// Server registration
instance := loadbalancer.ServiceInstance{
    Address: "192.168.1.10:50051",
    Weight:  100,
    Metadata: map[string]string{
        "capacity": "high",
    },
}
```

## Best Practices

1. **Health Checks**: Enable periodic health checks to remove unhealthy endpoints
2. **Jitter**: Enable jitter to prevent thundering herd when endpoints recover
3. **Weight Tuning**: Adjust weights based on server capacity and latency

## Monitoring

When Prometheus integration is enabled:

| Metric | Type | Description |
|--------|------|-------------|
| `rpc_lb_requests_total` | Counter | Total requests per endpoint |
| `rpc_lb_active_connections` | Gauge | Active connections per endpoint |

## Next Steps

- [Circuit Breaker](../architecture/circuit-breaker.md) — FAILURE containment
- [Observability](../guides/observability.md) — Metrics collection
