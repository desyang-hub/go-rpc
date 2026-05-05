# Rate Limiting

This guide covers configuring rate limiting for go-rpc services.

## Overview

Rate limiting protects services from overload by controlling request throughput.

## Supported Algorithms

| Algorithm | Use Case |
|-----------|----------|
| Token Bucket | Smooth traffic bursts |
| Sliding Window | Accurate rate tracking |
| Fixed Window | Simple implementation |

## Token Bucket Configuration

### YAML Configuration

```yaml
# config.yaml
ratelimit:
  enabled: true
  algorithm: token_bucket
  rate: 1000          # requests per second
  burst: 1500         # maximum burst size
  per_service: true   # per-service limits
  endservices:
    hello:
      rate: 500
      burst: 1000
    health:
      rate: 100
      burst: 200
```

### Go Configuration

```go
import "github.com/desyang-hub/go-rpc/pkg/limiter"

limiter := limiter.New(limiter.Config{
    Algorithm: limiter.TokenBucket,
    Rate:      1000,  // requests/second
    Burst:     1500,  // max burst
})

middleware := middleware.RateLimit(limiter).
    ErrorFunc(func(ctx context.Context, err error) error {
        return status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %v", err)
    }).
    ClientIdentifier(func(ctx context.Context) string {
        if md, ok := metadata.FromIncomingContext(ctx); ok {
            if ip := md.Get("x-forwarded-for"); len(ip) > 0 {
                return ip[0]
            }
        }
        return "unknown"
    })
```

## Fixed Window Configuration

### YAML Configuration

```yaml
ratelimit:
  enabled: true
  algorithm: fixed_window
  window: 60s         # time window
  limit: 3600         # requests per window
```

## Custom Rate Limit Response

```go
middleware := middleware.RateLimit(limiter).
    Code(func(ctx context.Context, err error) codes.Code {
        // Return status code for rate limited requests
        return codes.ResourceExhausted
    })
```

## Best Practices

1. **Start Conservative**: Begin with generous limits and tighten based on metrics
2. **Per-Service Limits**: Configure different limits per endpoint
3. **Monitor**: Use Prometheus metrics to track rate limit violations
4. **Graceful Degradation**: Return meaningful error messages to clients

## Monitoring

When Prometheus integration is enabled:

| Metric | Type | Description |
|--------|------|-------------|
| `ratelimit_violations_total` | Counter | Total rate limit violations |
| `ratelimit_requests_allowed_total` | Counter | Total allowed requests |

## Next Steps

- [Docker Deployment](../deployment/docker.md) — Deploy with rate limiting
- [Contributing](../contributing.md) — Contribute to the project
