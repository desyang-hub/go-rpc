# Core Components

This document describes the core components of the go-rpc framework.

## Overview

The go-rpc framework consists of several core components that work together to provide a complete RPC solution.

## Server (pkg/server)

The server component handles gRPC server lifecycle management.

```go
import "github.com/desyang/go-rpc/pkg/server"

// Create a new server with configuration
srv := server.NewServer().
    Address(":50051").
    MaxConcurrentStreams(100).
    KeepaliveTime(30 * time.Minute).
    AddMiddleware(middleware.Logging()).
    Build()
```

### Features

| Feature | Default | Configurable |
|---------|---------|--------------|
| Address Port | 50051 | Yes |
| Max Concurrent Streams | 100 | Yes |
| Keepalive Time | 30 min | Yes |
| Graceful Shutdown | 30s timeout | Yes |
| TLS | Optional | Yes |

### Server Builder Pattern

The server uses a builder pattern for fluent configuration:

```go
srv := server.NewServer().
    Address(":50051").
    MaxConcurrentStreams(100).
    KeepaliveTime(30 * time.Minute).
    KeepaliveTimeout(10 * time.Minute).
    AddMiddleware(middleware.Logging()).
    AddMiddleware(middleware.Auth()).
    Build()
```

## Client (pkg/client)

The client component manages connections to remote services.

```go
import "github.com/desyang/go-rpc/pkg/client"

cl := client.NewClient().
    Address("server:50051").
    ServiceName("my-service").
    DiscoveryRegistry(discovery).
    LoadBalancer(balancer).
    RetryPolicy(client.DefaultRetryPolicy()).
    CircuitBreaker(circuitBreaker).
    Build()
```

### Features

| Feature | Description |
|---------|-------------|
| **Connection Pooling** | Maintains pooled connections to all endpoints |
| **Retry** | Automatic retry with exponential backoff |
| **Circuit Breaker** | Prevents calls to failing services |
| **Discovery** | Dynamic endpoint resolution via registry |
| **Health Check** | Periodic health probes to endpoints |

## Middleware (pkg/middleware)

Middleware provides a pluggable interceptor chain for cross-cutting concerns.

### Server Middleware

```go
type UnaryServerInterceptor func(ctx context.Context, req interface{}, 
    info *UnaryServerInfo, handler UnaryHandler) (interface{}, error)
```

### Available Middleware

| Middleware | Purpose |
|------------|---------|
| `Logging()` | Structured request/response logging |
| `Auth(tokenFunc)` | Token-based authentication |
| `RateLimit(rps)` | Request rate limiting |
| `Timeout(dur)` | Per-request timeout |

## Protobuf Definitions (api/)

Service contracts are defined using Protocol Buffers proto3:

```protobuf
syntax = "proto3";

package myservice.v1;

option go_package = "github.com/your-org/my-service/api;genapi";

message HelloRequest {
  string name = 1;
}

message HelloResponse {
  string message = 1;
}

service HelloService {
  rpc Hello(HelloRequest) returns (HelloResponse);
}
```

### Four Call Modes

| Mode | Definition | Use Case |
|------|------------|----------|
| **Unary** | `rpc Hello(Request) returns (Response)` | Standard request/response |
| **Server Streaming** | `rpc GetMany(Request) returns (stream Response)` | Large result sets |
| **Client Streaming** | `rpc Upload(stream Request) returns (Response)` | Large file uploads |
| **Bidirectional** | `rpc Stream(stream Request) returns (stream Response)` | Chat, live updates |

## Next Steps

- [Getting Started](../getting-started/quick-start.md) — How to use the framework
- [Service Registration Guide](../guides/service-registration.md) — Configure service discovery
- [Load Balancing Setup](../guides/load-balancing.md) — Configure load balancing
