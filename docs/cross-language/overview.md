# Cross-Language Overview

Cross-language interoperability allows services written in different languages to communicate through a shared interface.

## Supported Languages

| Language | Status | Module |
|----------|--------|--------|
| Go | Production | `github.com/desyang/go-rpc/pkg/server` |
| Python | Production | `pip install go-rpc-client` |
| TypeScript | Beta | `npm install @desyang/go-rpc-client` |

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                         Interface Layer                          │
│                            (proto3)                              │
└─────────────────────────────────────────────────────────────────┘
                                │
              ┌─────────────────┼─────────────────┐
              │                 │                 │
              ▼                 ▼                 ▼
    ┌───────────────┐  ┌───────────────┐  ┌───────────────┐
    │  Go Server    │  │ Python Client │  │ TS Client     │
    │  (Service)    │  │ (Consumer)    │  │ (Consumer)    │
    └───────────────┘  └───────────────┘  └───────────────┘
```

## Interface Design Principles

1. **Schema-First**: All interfaces defined in `.proto` files
2. **Versioned**: Package names encode versions (e.g., `myservice.v1`)
3. **Strongly Typed**: Type definitions ensure consistency
4. **Auto-Generated**: Clients generated from schema

## Code Generation Tool

```bash
./bin/rpc-gen --proto=./api --output=./generated --languages=python,typescript
```

Generated output:
```
./generated/
├── python/
│   ├── client.py
│   ├── grpc_generated.py
│   └── config.py
├── typescript/
│   ├── client.ts
│   ├── grpc_generated.ts
│   └── config.ts
└── go/
    └── *.pb.go
```

## Next Steps

- [Python Client](python.md) — Python integration guide
- [TypeScript Client](typescript.md) — TypeScript integration guide
- [Code Generation](code-generation.md) — Using rpc-gen
