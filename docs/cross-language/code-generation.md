# Code Generation

This guide covers using the `rpc-gen` CLI tool to generate client code for multiple languages.

## Overview

`rpc-gen` is a command-line tool that generates type-safe client code from Protocol Buffer definitions (.proto files).

## Installation

```bash
# From source
go install github.com/desyang-hub/go-rpc/cmd/rpc-gen@latest

# Download binary
curl -fsSL https://github.com/desyang-hub/go-rpc/releases/download/v0.2.0/rpc-gen_linux_amd64 -o rpc-gen
chmod +x rpc-gen
```

## Basic Usage

```bash
rpc-gen --proto=./api --output=./generated --languages=go,python,typescript
```

## Command-Line Options

```
rpc-gen --help

Usage: rpc-gen [options]

Options:
  -p, --proto DIRECTORY        Path to proto files (required)
  -o, --output DIRECTORY       Output directory for generated code (required)
  -l, --languages STRING       Comma-separated language list (e.g., go,python,typescript)
  -w, --web                    Generate gRPC-Web client code
  --include-imports            Include imports for external proto dependencies
  -r, --recursive              Recursively find proto files
  --config PATH                Path to rpc-gen configuration file
  -h, --help                   Show help information
```

## Configuration File

Create an `rpc-gen.yaml` configuration file:

```yaml
# rpc-gen.yaml

# Proto files location
proto:
  paths:
    - ./api/proto
  recursive: true

# Output configuration
output:
  directory: ./generated

# Languages to generate
languages:
  - go
  - python
  - typescript

# Per-language options
options:
  go:
    package: myapi
    import_path: github.com/your-org/my-service/api

  python:
    module_prefix: generated
    async: true

  typescript:
    module_prefix: generated
    async: true
    webpack: false
```

## Generated Code Structure

### Go

```yaml
./generated/
└── go/
    └── myapi/
        ├── hello.pb.go
        └── hello_grpc.pb.go
```

### Python

```yaml
./generated/
└── python/
    ├── generated/
    │   ├── __init__.py
    │   ├── hello_pb2.py
    │   ├── hello_pb2_grpc.py
    │   └── client.py
    └── setup.py
```

### TypeScript

```yaml
./generated/
└── typescript/
    ├── src/
    │   ├── generated/
    │   │   ├── hello_pb.d.ts
    │   │   ├── hello_pb.js
    │   │   ├── hello_grpc.d.ts
    │   │   ├── hello_grpc.js
    │   │   └── client.ts
    │   └── index.ts
    ├── package.json
    └── tsconfig.json
```

## Plugin System

Extend `rpc-gen` with custom plugins:

```go
// plugin.go
package main

import "github.com/desyang-hub/go-rpc/pkg/generator"

// Register custom generator
generator.Register("java", func(g *generator.Generator) error {
    // Generate Java code from proto files
    return nil
})
```

## Makefile Integration

```makefile
.PHONY: generate
generate:
    rpc-gen --proto=./api --output=./generated --languages=go,python,typescript
```

## Next Steps

- [Python Client](python.md) — Using generated Python code
- [TypeScript Client](typescript.md) — Using generated TypeScript code
- [Docker Deployment](../deployment/docker.md) — Deploying with Docker
