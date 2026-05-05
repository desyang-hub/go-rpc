# Architecture Overview

This document describes the overall architecture of the go-rpc framework.

## System Architecture

go-rpc follows a layered architecture pattern, separating concerns into distinct layers:

```
┌─────────────────────────────────────────────────────────────────┐
│                         Application Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────────────────────┐ │
│  │ Go Server   │  │ Go Client   │  │ rpc-gen (CLI Tool)       │ │
│  └──────┬──────┘  └──────┬──────┘  └─────────────┬────────────┘ │
└─────────┼────────────────┼────────────────────────┼──────────────┘
          │                │                        │
┌─────────┼────────────────┼────────────────────────┼──────────────┐
│         │    RPC Engine  │         Plugin Layer    │              │
│         │                │                         │              │
│    ┌────┴────┐    ┌──────┴──────┐    ┌─────────────┴────────────┐ │
│    │ gRPC    │    │  Middleware  │    │  Discovery LB Circuit    │ │
│    │ Client/ │    │   Chain      │    │  Breaker Retry Timeout   │ │
│    │ Server  │    └─────────────┘    └──────────────────────────┘ │
│    └────┬────┘                                                    │
└─────┬───┴─────────────────────────────────────────────────────────┘
      │
┌─────┴─────────────────────────────────────────────────────────────┐
│                       Transport Layer                               │
│  ┌─────────┐  ┌──────────┐  ┌──────────┐  ┌────────────────────┐  │
│  │ HTTP/2  │  │ TLS/SSL  │  │ Keepalive│  │ Connection Pool    │  │
│  └─────────┘  └──────────┘  └──────────┘  └────────────────────┘  │
└───────────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. RPC Engine

The RPC Engine handles the core request lifecycle:

- **Client**: Manages gRPC connections, interceptors, and retry logic
- **Server**: Manages gRPC service registration, graceful shutdown, and health checks

### 2. Plugin Layer

Each plugin implements a well-defined interface:

| Plugin | Purpose | Interfaces |
|--------|---------|------------|
| **Middleware** | Request interception | `UnaryInterceptor`, `StreamInterceptor` |
| **Service Discovery** | Service registration and discovery | `Registry`, `Watcher` |
| **Load Balancer** | Client-side load distribution | `Balancer`, `AddressPicker` |
| **Circuit Breaker** | Failure containment | `CircuitBreaker`, `Fallback` |

### 3. Transport Layer

The transport layer provides connection management and security:

- **Connection Pool**: Reusable gRPC connections with health checking
- **TLS/SSL**: End-to-end encryption support
- **Keepalive**: Periodic connection health probes
- **TLS Authentication**: Mutual TLS support for service-to-service auth

## Request Flow

```
┌──────────┐     ┌──────────────┐     ┌──────────────┐     ┌──────────┐
│  Client  │────>│Load Balancer │────>│  Server      │     │  Service │
│  (App)   │     │(Pick Node)   │     │(gRPC Server) │────>│(Handler) │
└──────────┘     └──────────────┘     └──────────────┘     └──────────┘
     │                                              │
     │              ┌──────────────┐                │
     │<─────────────│Response      │<───────────────┘
     │              └──────────────┘
     │
     │          ┌──────────────┐     ┌──────────────┐
     │<─────────│Circuit      │────<│Server Health │
     │          │Breaker      │     │Monitor       │
     │          └──────────────┘     └──────────────┘
```

## Service Discovery

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│   Server     │────>│  Registry    │     │   Client     │
│  (Registrar) │     │(Consul/etcd) │<───>│(Watcher)     │
└──────────────┘     └──────────────┘     └──────────────┘
```

The server registers itself with a registry (Consul or etcd), and the client watches registry changes to build its service list dynamically.

## Scalability Considerations

- **Horizontal Scaling**: Multiple server instances register via the same registry
- **Connection Pooling**: Clients maintain persistent connections to all discovered nodes
- **Backpressure**: gRPC flow control prevents overwhelming downstream services
- **Graceful Degradation**: Circuit breaker prevents cascade failures

## Next Steps

- [Core Components](architecture/core-components.md) — Detailed component documentation
- [Service Discovery](architecture/service-discovery.md) — Registry integrations
- [Load Balancing](architecture/load-balancing.md) — Client-side strategies
- [Circuit Breaker](architecture/circuit-breaker.md) — Fault tolerance mechanisms
