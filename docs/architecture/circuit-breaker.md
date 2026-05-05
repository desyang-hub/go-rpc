# Circuit Breaker

This document describes the circuit breaker pattern implementation in go-rpc.

## Overview

The circuit breaker pattern prevents cascade failures by stopping requests to failing services and allowing recovery time.

## States

```
                    ┌──────────────┐
        success     │   CLOSED     │
     ┌─────────────>│   stable     │
     │              │   ~healthy   │
     │              └──────────────┘
     │                       │
     │              failure  │
     │               exceeds │
     │               threshold
     │                       ▼
     │              ┌──────────────┐
     └─────────────<│   OPEN       │
                    │   tripped    │
                    │   ~failing   │
                    └──────────────┘
                           │
                    timeout
                    elapsed
                           ▼
              ┌──────────────────┐
              │  HALF-OPEN       │
              │  ~testing        │
              └──────────────────┘
                      │       │
               success   failure
                      │       │
                      ▼       ▼
                 ┌─────────┐  ┌─────────┐
                 │ CLOSED  │  │ OPEN    │
                 └─────────┘  └─────────┘
```

## States

| State | Behavior |
|-------|----------|
| **CLOSED** |正常工作，请求正常通过。失败超过阈值时变为 OPEN |
| **OPEN** | 阻止所有请求，直接返回错误。超时后变为 HALF-OPEN |
| **HALF-OPEN** | 允许有限数量的试探请求。成功则 CLOSED，失败则 OPEN |

## Configuration

```go
import "github.com/desyang/go-rpc/pkg/circuitbreaker"

cb := circuitbreaker.New("my-service",
    circuitbreaker.WithFailureThreshold(5),      // 连续失败 5 次打开
    circuitbreaker.WithRecoveryTimeout(30 * time.Second), // 30s 后尝试
    circuitbreaker.WithHalfOpenAttempts(3),       // 半开状态允许 3 次试探
)
```

## YAML Configuration

```yaml
# config.yaml
circuit_breaker:
  enabled: true
  service: my-service
  failure_threshold: 5
  recovery_timeout: 30s
  half_open_attempts: 3
  fallback_function: "handleFallback"
```

## Fallback Mechanism

Define fallback logic when the circuit breaker is open:

```go
import "github.com/desyang/go-rpc/pkg/circuitbreaker"

fallback := func() (interface{}, error) {
    // Return cached data or default response
    return CachedResponse(), nil
}

cb := circuitbreaker.New("my-service",
    circuitbreaker.WithFailureThreshold(5),
    circuitbreaker.WithRecoveryTimeout(30 * time.Second),
    circuitbreaker.WithFallback(fallback),
)
```

## Example: Circuit Breaker in Client

```go
cl := client.NewClient().
    Address("server:50051").
    CircuitBreaker(cb).
    RetryPolicy(client.DefaultRetryPolicy(
        3,  // attempts
        100 * time.Millisecond,
        2 * time.Second,
    )).
    Build()
```

## Metrics Exposure

When Prometheus integration is enabled, circuit breaker exposes:

| Metric | Type | Labels |
|--------|------|--------|
| `rpc_circuit_breaker_state` | Gauge | `service`, `state` |
| `rpc_circuit_breaker_failures_total` | Counter | `service` |
| `rpc_circuit_breaker_success_total` | Counter | `service` |
| `rpc_circuit_breaker_opened_total` | Counter | `service` |

## Error Probabilities

### When Circuit Is CLOSED

| Scenario | Behavior |
|----------|----------|
| Normal call | Request goes to server |
| Server error | Failure counter increments |
| Failure threshold reached | Circuit opens |

### When Circuit Is OPEN

| Scenario | Behavior |
|----------|----------|
| Normal call | Returns error immediately |
| Fallback defined | Returns fallback result |
| After timeout | Transitions to HALF-OPEN |

### When Circuit Is HALF-OPEN

| Scenario | Behavior |
|----------|----------|
| Successful call | Circuit closes (recovery) |
| Failed call | Circuit reopens |
| Probe attempts exhausted | Circuit closes (recovery) |

## Next Steps

- [Service Registration Guide](../guides/service-registration.md) — Service discovery integration
- [Observability Guide](../guides/observability.md) — Metrics collection
