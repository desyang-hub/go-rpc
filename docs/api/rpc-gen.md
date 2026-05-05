# rpc-gen CLI

This document provides API reference documentation for the rpc-gen code generation tool.

## Overview

rpc-gen is a CLI tool that generates type-safe client code from Protocol Buffer definitions for multiple languages (Go, Python, TypeScript).

## Command-Line Reference

```
rpc-gen [options]

Options:
  -p, --proto DIRECTORY    Path to proto files (required)
  -o, --output DIRECTORY   Output directory for generated code (required)
  -l, --languages STRING   Comma-separated language list (default: go)
                           Supported: go, python, typescript
  -w, --web                Generate gRPC-Web client code (TypeScript only)
  -r, --recursive          Recursively find proto files
  --include-imports        Include imports for external proto dependencies
  --config PATH            Path to rpc-gen configuration file
  -v, --version            Print version information
  -h, --help               Show help information
```

## Version Information

```
rpc-gen version
```

Output:

```
rpc-gen version 0.2.0
Build timestamp: 2025-01-15T10:30:00Z
Go version: 1.22.0
```

## Configuration

### Command-Line Options

```bash
# Generate code for all supported languages
rpc-gen --proto=./api --output=./generated --languages=go,python,typescript

# Generate only Python client
rpc-gen --proto=./api --output=./generated --languages=python

# Generate with gRPC-Web support
rpc-gen --proto=./api --output=./generated --languages=typescript --web

# Recursive proto file discovery
rpc-gen --proto=./api --output=./generated --recursive
```

### Configuration File (rpc-gen.yaml)

```yaml
# rpc-gen.yaml
proto:
  paths:
    - ./api/proto
  recursive: true

output:
  directory: ./generated

languages:
  - go
  - python
  - typescript

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
    web: true
```

Using configuration file:

```bash
rpc-gen --config ./rpc-gen.yaml
```

## Output Structure

### Go Output

```yaml
./generated/
└── go/
    └── myapi/
        ├── hello.pb.go
        └── hello_grpc.pb.go
```

### Python Output

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

### TypeScript Output

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

Extend rpc-gen with custom code generators:

```go
package main

import _ "github.com/your-org/rpc-gen-plugins/java"
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid arguments |
| 3 | I/O error |
| 4 | Generation error |