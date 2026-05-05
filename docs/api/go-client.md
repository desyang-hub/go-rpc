# Go Client API Reference

This document provides API reference documentation for the go-rpc Go client package.

## Package: pkg/client

### Overview

The client package provides a high-level gRPC client with built-in features like connection pooling, retry logic, circuit breaking, and service discovery integration.

## Client Configuration

### Config Struct

```go
type Config struct {
    Address           string            // Server address(es)
    ServiceName       string            // Service discovery name
    Timeout           time.Duration     // Request timeout (default: 30s)
    MaxConnsPerHost   int               // Max connections per host (default: 10)
    RetryPolicy       RetryPolicy       // Retry configuration
    CircuitBreaker    CircuitBreaker    // Circuit breaker
    Discovery         ServiceDiscovery  // Service discovery implementation
    LoadBalancer      LoadBalancer      // Load balancer strategy
    TLS               *tls.Config       // TLS configuration
    Logger            *zap.Logger       // Structured logger
    Tracer            *opentelemetry.Tracer // OpenTelemetry tracer
}
```

### Retry Policy

```go
type RetryPolicy struct {
    MaxRetries     int           // Maximum retry attempts (default: 3)
    InitialDelay   time.Duration // First retry delay (default: 100ms)
    MaxDelay       time.Duration // Maximum retry delay (default: 10s)
    BackoffFactor  float64       // Backoff multiplier (default: 2.0)
    RetryableCodes []codes.Code  // gRPC codes that trigger retry
}
```

### Building a Client

```go
import "github.com/desyang/go-rpc/pkg/client"

// Simple client
cl := client.NewClient().
    Address("localhost:50051").
    Build()

// Full-featured client
cl := client.NewClient().
    Address("192.168.1.10:50051,192.168.1.11:50051").
    ServiceName("my-service").
    Timeout(30 * time.Second).
    RetryPolicy(client.DefaultRetryPolicy()).
    CircuitBreaker(circuitBreaker).
    Build()
```

## Client Methods

### Dial()

```go
func (c *Client) Dial(ctx context.Context) error
```

Establishes a connection to the gRPC server.

```go
ctx := context.Background()
if err := cl.Dial(ctx); err != nil {
    log.Fatalf("Failed to connect: %v", err)
}
```

### Close()

```go
func (c *Client) Close() error
```

Closes the client connection and releases all resources.

```go
defer cl.Close()
```

### UnaryCall()

```go
func (c *Client) UnaryCall(ctx context.Context, method string, req interface{}, 
    opts ...grpc.CallOption) (interface{}, error)
```

Makes a unary RPC call.

```go
request := &HelloRequest{Name: "World"}
response := &HelloResponse{}

if err := cl.UnaryCall(ctx, "/HelloService/SayHello", request, grpc.Header(&metadata)); err != nil {
    log.Fatal(err)
}
```

### ServerStream()

```go
func (c *Client) ServerStream(ctx context.Context, method string, req interface{}, 
    opts ...grpc.CallOption) (*grpc.ClientStream, error)
```

Makes a server-streaming RPC call.

```go
stream, err := cl.ServerStream(ctx, "/HelloService/Broadcast", &BroadcastRequest{
    RoomId: "general",
})
if err != nil {
    log.Fatal(err)
}

for {
    msg, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(msg)
}
```

### ClientStream()

```go
func (c *Client) ClientStream(ctx context.Context, method string, opts ...grpc.CallOption) (*grpc.ClientStream, error)
```

Makes a client-streaming RPC call.

```go
stream := cl.ClientStream(ctx, "/UploadService/Upload")
for i := 0; i < 100; i++ {
    chunk := &UploadChunk{Index: i, Data: makeChunk(i)}
    if err := stream.Send(chunk); err != nil {
        log.Fatal(err)
    }
}
response, err := stream.CloseAndRecv()
```

## Next Steps

- [Go Server API](go-server.md) — Server-side API reference
- [rpc-gen CLI](rpc-gen.md) — Code generation tool
