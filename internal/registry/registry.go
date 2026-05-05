// Package registry provides service registration and discovery interfaces.
//
// The registry package defines a pluggable interface that can be implemented
// with various backends (Consul, etcd, etc.). Service instances register
// themselves with healthy status on startup and deregister on shutdown.
//
// Usage:
//
//	// Register a service instance
//	cfg := &ConsulConfig{
//	    Address: "127.0.0.1:8500",
//	    ServiceName: "go-rpc.server",
//	    ServiceID:   "go-rpc.server-1",
//	}
//	b, _ := NewConsulBuilder(cfg)
//	reg := b.Build()
//	reg.Register(ctx, server.ServiceInfo{
//	    Name:     "go-rpc.server",
//	    Addr:     "127.0.0.1:50051",
//	    Metadata: map[string]string{"protocol": "gRPC"},
//	})
package registry

import (
	"context"
	"encoding/json"
	"time"
)

// ServiceInfo contains information about a registered service instance.
type ServiceInfo struct {
	Name     string
	Addr     string
	Metadata map[string]string
	TTL      time.Duration
	Healthy  bool
}

// ToJSON serializes ServiceInfo to JSON.
func (s ServiceInfo) ToJSON() string {
	data, _ := json.Marshal(s)
	return string(data)
}

// Registry provides service registration and discovery.
type Registry interface {
	// Register registers a service instance with the registry.
	Register(ctx context.Context, service ServiceInfo) error

	// Deregister removes a service instance from the registry.
	Deregister(ctx context.Context, service ServiceInfo) error

	// Watch returns a channel that receives updates when service instances
	// are registered or deregistered.
	Watch(ctx context.Context) (<-chan ServiceUpdate, error)

	// Close releases any resources held by the registry.
	Close() error
}

// ServiceUpdate represents a change in service registration.
type ServiceUpdate struct {
	Service ServiceInfo
	Type    UpdateType
}

// UpdateType indicates the type of service change.
type UpdateType int

const (
	UpdateTypeRegister UpdateType = iota + 1
	UpdateTypeDeregister
	UpdateTypeUpdate
)

// Config contains registry configuration.
type Config struct {
	ServiceName string
	ServiceID   string
	TTL         time.Duration
	HealthCheck string
	Metadata    map[string]string
}
