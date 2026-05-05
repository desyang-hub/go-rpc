# Go Server API Reference

This document provides API reference documentation for the go-rpc Go server package.

## Package: pkg/server

### Overview

The server package provides a high-performance gRPC server with built-in features like TLS support, keepalive configuration, and middleware integration.

## Server Configuration

### Config Struct

```go
type Config struct {
    Address          string        // gRPC server address (default: ":50051")
    Port             int           // gRPC server port (default: 50051)
    MaxConcurrentStreams int     // gRPC max concurrent streams (default: 100)
    MaxConnectionAge time.Duration // Maximum connection age before shutdown
    KeepaliveTime    time.Duration // Keepalive probe interval (default: 30m)
    KeepaliveTimeout time.Duration // Keepalive timeout (default: 10m)
    TLS              *tls.Config   // TLS configuration
    Logger           *zap.Logger   // Structured logger
    Metrics          *prometheus.Metrics // Prometheus metrics instance
    Tracer           *opentelemetry.Tracer // OpenTelemetry tracer
}
```

### Building a Server

```go
import "github.com/desyang/go-rpc/pkg/server"

srv := server.NewServer()
srv.Configure(server.Config{
    Address: ":50051",
})

// Or use the builder pattern
srv := server.NewServer().
    Address(":50051").
    MaxConcurrentStreams(100).
    KeepaliveTime(30 * time.Minute).
    KeepaliveTimeout(10 * time.Minute).
    Build()
```

## Server Methods

### Start()

```go
func (s *Server) Start(ctx context.Context) error
```

Starts the gRPC server with the given context. The server will listen for shutdown signals.

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()
srv.Start(ctx)
```

### Shutdown()

```go
func (s *Server) Shutdown(timeout time.Duration) error
```

Gracefully stops the server, allowing in-progress requests to complete.

```go
srv.Shutdown(30 * time.Minute)
```

## Registration Methods

### RegisterGRPCService()

```go
func (s *Server) RegisterGRPCService(
    desc *grpc.ServiceDesc, 
    impl interface{}, 
    opts ...grpc.ServerOption, 
)
```

Registers a gRPC service with the server.

```go
srv.RegisterGRPCService(
    &HelloService_ServiceDesc, 
    &HelloServiceImpl{},
)
```

### RegisterHealthCheck()

```go
func (s *Server) RegisterHealthCheck(handler *health.StatusCheck)
```

Registers a health check endpoint.

```go
import "github.com/desyang/go-rpc/pkg/health"

handler := health.NewStatusCheck()
srv.RegisterHealthCheck(handler)
```

## Helper Methods

### GRPCServer()

```go
func (s *Server) GRPCServer() *grpc.Server
```

Returns the underlying gRPC server instance for custom configurations.

## Middleware Integration

### AddMiddleware()

```go
func (s *Server) AddMiddleware(interceptor grpc.UnaryServerInterceptor)
```

Adds a unary server interceptor to the server.

```go
srv.AddMiddleware(middleware.Logging())
srv.AddMiddleware(middleware.Auth())
srv.AddMiddleware(middleware.RateLimit(limiter))
```

## Next Steps

- [Go Client API](go-client.md) — Client-side API reference
- [rpc-gen CLI](rpc-gen.md) — Code generation tool
