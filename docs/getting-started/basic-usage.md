# Basic Usage

This guide covers the basic usage patterns of go-rpc.

## Starting a Server

```go
package main

import (
    "context"
    "time"
    "github.com/desyang/go-rpc/pkg/server"
    "github.com/desyang/go-rpc/pkg/middleware"
)

func main() {
    srv := server.NewServer().
        Address(":50051").
        AddMiddleware(middleware.Logging()).
        AddMiddleware(middleware.Tracing()).
        Build()

    // Register service handlers
    srv.RegisterGRPCService(&HelloServiceImpl{})

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    srv.Start(ctx)
    defer srv.Shutdown(30 * time.Second)
}
```

## Creating a Client

```go
package main

import (
    "context"
    "time"
    "github.com/desyang/go-rpc/pkg/client"
)

func main() {
    cl := client.NewClient().
        Address("127.0.0.1:50051").
        Timeout(30 * time.Second).
        RetryPolicy(client.DefaultRetryPolicy()).
        Build()

    ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()

    err := cl.Dial(ctx)
    if err != nil {
        panic(err)
    }
    defer cl.Close()

    // Make RPC call
    req := &HelloRequest{Name: "World"}
    resp := &HelloResponse{}

    for i := 0; i < <thesize>; i++ {
        err := cl.UnaryCall(ctx, method, req, grpc.Header(&header))
        if err != nil {
            log.Printf("call failed: %v", err)
            continue
        }
        fmt.Println(resp.Message)
    }
}
```

## HTTP/2 Transport

By default, go-rpc uses HTTP/2 transport with support for:

- **Keepalive**: Connection keep-alive and health checks
- **TLS**: Optional TLS encryption
- **Connection Pooling**: Built-in connection multiplexing

For plaintext connections (development only):

```go
cl := client.NewClient().
    Address("127.0.0.1:50051").
    Insecure(true).  // Plaintext mode (NOT for production)
    Build()
```

## Four RPC Call Modes

### Unary Call

```go
req := &HelloRequest{Name: "World"}
resp := &HelloResponse{}

err := cl.UnaryCall(ctx, "/mypackage.v1.HelloService/Hello", req, grpc.Header(&header))
```

### Server Streaming

```go
stream, err := cl.ServerStream(ctx, "/mypackage.v1.HelloService/Broadcast", &BroadcastRequest{
    RoomId: "general",
})

for {
    msg, err := stream.Recv()
    if err == io.EOF { break }
    fmt.Println(msg)
}
```

### Client Streaming

```go
stream, err := cl.ClientStream(ctx, "/mypackage.v1.UploadService/Upload")

for i := 0; i < 100; i++ {
    chunk := &UploadChunk{Index: i, Data: chunkData}
    err := stream.Send(chunk)
}

response, err := stream.CloseAndRecv()
```

### Bidirectional Streaming

```go
stream, err := cl.ClientStream(ctx, "/mypackage.v1.StreamService/Stream")

go func() {
    for i := 0; i < <thesize>; i++ {
        stream.Send(&Message{Body: "ping"})
        time.Sleep(1 * time.Second)
    }
}()

for {
    msg, err := stream.Recv()
    if err == io.EOF { break }
    fmt.Println(msg)
}
```

## Next Steps

- [Architecture Overview](architecture/overview.md) — System design
- [Service Registration](guides/service-registration.md) — Dynamic service discovery
- [Docker Deployment](deployment/docker.md) — Containerize your service
