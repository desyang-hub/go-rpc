# go-rpc Documentation

Enterprise-grade cross-language RPC framework built for production environments.

## Overview

go-rpc is a high-performance RPC framework based on gRPC and Protocol Buffers, designed for building distributed systems with first-class support for multiple programming languages.

## Key Features

| Feature | Description |
|---------|-------------|
| **gRPC Standard** | Based on Protocol Buffers and HTTP/2, supporting all four call modes |
| **Service Discovery** | Consul and etcd backends, switchable via configuration |
| **Load Balancing** | Round-robin, weighted round-robin, least connections, consistent hashing |
| **Circuit Breaker** | Google SRE algorithm with custom degradation logic |
| **Observability** | OpenTelemetry tracing + Prometheus metrics |
| **Cross-Language** | Auto-generate Python/TypeScript client code via `rpc-gen` |
| **Pluggable Architecture** | Interceptors, discovery backends, and load balancers are all swappable |

## Architecture

![Architecture](assets/architecture.png)

See the [Architecture Overview](architecture/overview.md) for detailed design and component interactions.

## Quick Links

- **Getting Started**: [Quick Start](getting-started/quick-start.md) · [Installation](getting-started/installation.md)
- **Guides**: [Service Registration](guides/service-registration.md) · [Observability](guides/observability.md)
- **Cross-Language**: [Python Client](cross-language/python.md) · [TypeScript Client](cross-language/typescript.md)
- **API Reference**: [Go Server](api/go-server.md) · [rpc-gen CLI](api/rpc-gen.md)

## Repository

[GitHub](https://github.com/desyang/go-rpc)
