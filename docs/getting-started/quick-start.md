# Quick Start

Learn how to use go-rpc in your project.

## Prerequisites

- Go 1.22+
- Docker & Docker Compose (for running examples)
- protoc compiler + plugins

## Installation

```bash
go get github.com/desyang/go-rpc/pkg/...
```

## Basic Usage

### 1. Define your service in proto

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

### 2. Generate Go code

```bash
protoc --go_out=. --go-grpc_out=. api/hello.proto
```

### 3. Implement Server

```go
import "github.com/desyang/go-rpc/pkg/server"
import "github.com/desyang/go-rpc/pkg/middleware"

func main() {
    srv := server.NewServer().
        Address(":50051").
        KeepaliveTime(30 * time.Minute).
        AddMiddleware(middleware.Logging()).
        Build()
    
    // Register your service
    pb.RegisterHelloServiceServer(srv.GRPCServer(), &HelloServiceImpl{})
    
    srv.Start(context.Background())
    defer srv.Shutdown(context.Background())
}
```

### 4. Implement Client

```go
import "github.com/desyang/go-rpc/pkg/client"

func main() {
    cl := client.NewClient().
        Address("127.0.0.1:50051").
        RetryPolicy(client.DefaultRetryPolicy()).
        Build()
    
    if err := cl.Dial(context.Background()); err != nil {
        log.Fatal(err)
    }
    defer cl.Close()
    
    // Use your generated service stub
}
```

## Run Example

```bash
cd examples/go-server
docker compose up

cd ../python-client
pip install -r requirements.txt
python client.py
```
