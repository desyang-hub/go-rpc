# Installation

## Prerequisites

### Go

```bash
# Verify Go 1.22+ is installed
go version
```

### Protocol Buffers

```bash
# Install protoc
brew install protobuf   # macOS
apt install protobuf-compiler  # Ubuntu

# Install Go plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Install gRPC-Gateway (optional)
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
```

### Docker & Docker Compose

Required for running examples and development environment.

## Install go-rpc Library

Add to your Go module:

```bash
go get github.com/desyang/go-rpc/pkg/server
go get github.com/desyang/go-rpc/pkg/client
go get github.com/desyang/go-rpc/pkg/middleware
```

Or add to go.mod directly:

```go
require (
    github.com/desyang/go-rpc v0.2.0
)
```

## Install rpc-gen CLI

```bash
go install github.com/desyang/go-rpc/cmd/rpc-gen@latest
```

Verify installation:

```bash
rpc-gen version
```
