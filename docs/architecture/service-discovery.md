# Service Discovery

This document describes the service discovery implementation in go-rpc.

## Overview

Service discovery enables dynamic registration and discovery of server instances, allowing clients to find and connect to services without hardcoded addresses.

## Architecture

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│   Server A   │───────>│  Consul/etcd │<──────│   Client     │
│  Instance 1  │       │  Registry    │──────>│  Watcher     │
├──────────────┤       └──────────────┘       └──────────────┘
│   Server B   │───────>│                │──────>│  Endpoint    │
│  Instance 2  │       │                │       │  List        │
├──────────────┤       │                │       └──────────────┘
│   Server C   │───────>│                │              │
└──────────────┘       └──────────────┘              │
                                                       │
                                               ┌───────┴───────┐
                                               │  Load Balancer │
                                               └───────────────┘
```

## Interfaces

### Registry Interface

All registry implementations must implement the `Registry` interface:

```go
type Registry interface {
    // Register registers a service instance with the registry
    Register(ctx context.Context, service ServiceInstance) error
    
    // Deregister removes a service instance from the registry
    Deregister(ctx context.Context, service ServiceInstance) error
    
    // Discover returns all healthy instances for a given service
    Discover(ctx context.Context, service string) ([]ServiceInstance, error)
    
    // Watch returns a channel that notifies of service changes
    Watch(ctx context.Context, service string) (<-chan *WatchEvent, error)
}

type ServiceInstance struct {
    Address    string
    Weight     int
    Metadata   map[string]string
    Healthy    bool
}

type WatchEvent struct {
    Type      WatchEventType
    Instances []ServiceInstance
}

type WatchEventType int

const (
    WatchAdd WatchEventType = iota
    WatchDelete
    WatchUpdate
)
```

## Consul Implementation

### Configuration

```yaml
# config.yaml
discovery:
  backend: consul
  consul:
    address: "localhost:8500"
    scheme: "http"
    datacenter: "dc1"
    persist: true
```

### Go Usage

```go
import "github.com/desyang-hub/go-rpc/pkg/discovery/consul"

registry := consul.NewRegistry(consul.Config{
    Address:  "localhost:8500",
    Scheme:   "http",
    Datacenter: "dc1",
})

instance := consul.ServiceInstance{
    Address: "127.0.0.1:50051",
    Weight:  100,
    Metadata: map[string]string{
        "version": "1.0.0",
    },
}

registry.Register(ctx, instance)
defer registry.Deregister(ctx, instance)

// Watch for changes
watchCh, err := registry.Watch(ctx, "my-service")
if err != nil {
    log.Fatal(err)
}

go func() {
    for event := range watchCh {
        fmt.Printf("Service changed: %d instances\n", len(event.Instances))
    }
}()
```

## etcd Implementation

### Configuration

```yaml
# config.yaml
discovery:
  backend: etcd
  etcd:
    endpoints:
      - "localhost:2379"
      - "localhost:2380"
    keyPrefix: "/services"
```

### Go Usage

```go
import "github.com/desyang-hub/go-rpc/pkg/discovery/etcd"

registry := etcd.NewRegistry(etcd.Config{
    Endpoints: []string{"localhost:2379", "localhost:2380"},
    KeyPrefix: "/services",
})

instance := etcd.ServiceInstance{
    Address: "127.0.0.1:50051",
    Weight:  100,
}

// List & Watch for live updates
watchCh, err := registry.ListAndWatch(ctx, "my-service")
```

## Registration/Discovery Flow

```
Server Startup                    Client Connection
────────────────                  ────────────────

 1. Resolve config
 2. Create registry impl
 3. Create service instance
 4. Register with registry                    1. Resolve config
                                         │    2. Create registry impl
                                         │    3. Discover instances
                                         
 5. Health check self
 6. Listen for shutdown signals              4. Select endpoint
                                              5. Dial connection
 7. Start server                             6. Create client

                                              7. Watch registry updates
                                              8. Refresh endpoint list
```

## Best Practices

1. **Health Checks**: Always register health check endpoints (TCP or HTTP)
2. **TTL**: Set appropriate TTLs for auto-deregistration on crash
3. **Metadata**: Include useful metadata (version, region, zone) for routing decisions
4. **Watch Reconnection**: Implement automatic watch reconnection on failures
5. **Caching**: Cache discovered instances locally with short TTL

## Next Steps

- [Load Balancing Setup](../guides/load-balancing.md) — Configure load balancing strategies
- [Docker Deployment](../deployment/docker.md) — Deploy with Docker
