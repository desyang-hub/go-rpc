# Load Balancing

This document describes the load balancing strategies available in go-rpc.

## Overview

Client-side load balancing distributes requests across multiple server instances, ensuring even utilization and fault tolerance.

## Strategies

### Round Robin (Default)

Distributes requests round-robin across all available instances.

```go
import "github.com/desyang/go-rpc/pkg/loadbalancer"

balancer := loadbalancer.NewRoundRobin()
```

### Weighted Round Robin

Distributes requests based on configured weights, allowing traffic shaping.

```go
balancer := loadbalancer.NewWeightedRoundRobin(map[string]int{
    "192.168.1.10:50051": 100,
    "192.168.1.11:50051": 50,
    "192.168.1.12:50051": 25,
})
```

### Least Connections

Routes requests to the instance with the fewest active connections. Best for uneven workloads.

```go
balancer := loadbalancer.NewLeastConnections()
```

### Consistent Hashing

Routes requests with the same hash key to the same instance. Useful for session affinity.

```go
balancer = loadbalancer.NewConsistentHashing(func(request interface{}) string {
    // Extract user ID for session affinity
    if req, ok := request.(*HelloRequest); ok {
        return req.UserId
    }
    return ""
})
```

## Interface Definition

```go
type Balancer interface {
    // Pick selects an endpoint from the available instances
    Pick(ctx context.Context, instances []ServiceInstance, request interface{}) (ServiceInstance, error)
    
    // Update updates the available instances
    Update(instances []ServiceInstance) error
}
```

## Configuration

```yaml
# config.yaml
loadbalancer:
  strategy: weighted_round_robin
  health_checks:
    enabled: true
    interval: 10s
    timeout: 3s
  selection:
    jitter: true  # Add random jitter to prevent thundering herd
```

## Configuration Matrix

| Strategy | Best For | Session Affinity | Failover |
|----------|----------|------------------|----------|
| Round Robin | Uniform workloads | No | Automatic |
| Weighted RR | Heterogeneous servers | No | Automatic |
| Least Connections | Variable latency | No | Automatic |
| Consistent Hash | Session-based | Yes | Partial |

## Example Usage

```go
import "github.com/desyang/go-rpc/pkg/client"
import "github.com/desyang/go-rpc/pkg/loadbalancer"

balancer := loadbalancer.NewRoundRobin(
    loadbalancer.WithHealthCheck(10*time.Second),
    loadbalancer.WithJitter(true),
)

cl := client.NewClient().
    LoadBalancer(balancer).
    Build()
```

## Next Steps

- [Circuit Breaker](architecture/circuit-breaker.md) — Circuit breaker integration
- [Rate Limiting Guide](guides/rate-limiting.md) — Configure rate limiting
