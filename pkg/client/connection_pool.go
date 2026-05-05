// Package client provides a connection pool for managing multiple gRPC connections.
//
// The connection pool maintains active connections to multiple service instances
// and handles connection lifecycle management.
//
// When used with service discovery, the pool automatically updates available
// connections when service instances are added or removed.
package client

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpckeepalive "google.golang.org/grpc/keepalive"

	"github.com/desyang-hub/go-rpc/internal/loadbalancer"
)

// ConnConfig contains connection configuration for a single instance.
type ConnConfig struct {
	DialTimeout      time.Duration
	KeepalivePeriod  time.Duration
	KeepaliveTimeout time.Duration
	TLSConfig        *tls.Config
	Credentials      credentials.TransportCredentials
}

// DefaultConnectionConfig returns default connection configuration.
func DefaultConnectionConfig() ConnConfig {
	return ConnConfig{
		DialTimeout:      10 * time.Second,
		KeepalivePeriod:  30 * time.Second,
		KeepaliveTimeout: 10 * time.Second,
	}
}

// ConnectionPool maintains and manages multiple gRPC connections.
//
// The pool is designed to work with service discovery:
// 1. Service instances are discovered via registry watch
// 2. The pool creates/updates connections based on instance changes
// 3. A load balancer distributes requests across available connections
type ConnectionPool struct {
	config      ConnConfig
	lb          loadbalancer.Strategy
	connections map[string]*grpc.ClientConn
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewConnectionPool creates a new connection pool.
//
// Available connections are managed by lb (load balancer).
func NewConnectionPool(ctx context.Context, cfg ConnConfig, lb loadbalancer.Strategy) *ConnectionPool {
	if cfg.DialTimeout == 0 {
		cfg.DialTimeout = 10 * time.Second
	}

	pool := &ConnectionPool{
		config:      cfg,
		lb:          lb,
		connections: make(map[string]*grpc.ClientConn),
	}

	pool.ctx, pool.cancel = context.WithCancel(ctx)

	return pool
}

// UpdateInstances processes service discovery updates to add/remove connections.
func (p *ConnectionPool) UpdateInstances(ctx context.Context, updates <-chan ServiceUpdate) {
	go func() {
		for {
			select {
			case <-p.ctx.Done():
				return
			case <-ctx.Done():
				return
			case upd, ok := <-updates:
				if !ok {
					return
				}

				switch upd.Type {
		case UpdateTypeRegister:
				p.AddInstance(upd.Service.Addr)
		case UpdateTypeDeregister:
				p.RemoveInstance(upd.Service.Addr)
		}
			}
		}
	}()
}

// AddInstance adds a new instance to the pool and creates a connection.
func (p *ConnectionPool) AddInstance(addr string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if _, exists := p.connections[addr]; exists {
		return
	}

	if p.lb != nil {
		p.lb.AddInstance(loadbalancer.Instance{Addr: addr})
	}

	p.connections[addr] = nil
}

// RemoveInstance removes an instance from the pool and closes its connection.
func (p *ConnectionPool) RemoveInstance(addr string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn, exists := p.connections[addr]; exists {
		if conn != nil {
			conn.Close()
		}
		delete(p.connections, addr)
	}

	if p.lb != nil {
		p.lb.RemoveInstance(addr)
	}
}

// GetInstance attempts to establish a connection to the instance.
func (p *ConnectionPool) GetInstance(addr string) (*grpc.ClientConn, error) {
	p.mu.Lock()
	conn, exists := p.connections[addr]
	p.mu.Unlock()

	if !exists {
		return nil, fmt.Errorf("instance not found: %s", addr)
	}

	if conn == nil {
		if err := p.createConnection(addr); err != nil {
			return nil, fmt.Errorf("failed to create connection to %s: %w", addr, err)
		}

		p.mu.Lock()
		conn = p.connections[addr]
		p.mu.Unlock()
	}

	return conn, nil
}

func (p *ConnectionPool) createConnection(addr string) error {
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(grpckeepalive.ClientParameters{
			Time:                p.config.KeepalivePeriod,
			Timeout:             p.config.KeepaliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithTimeout(p.config.DialTimeout),
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	p.mu.Lock()
	p.connections[addr] = conn
	p.mu.Unlock()

	return nil
}

// Next returns the next instance from the load balancer and creates a connection.
func (p *ConnectionPool) Next() (*grpc.ClientConn, error) {
	instance, ok := p.lb.Next()
	if !ok {
		return nil, fmt.Errorf("no available instances")
	}

	conn, err := p.GetInstance(instance.Addr)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance %s: %w", instance.Addr, err)
	}

	return conn, nil
}

// GetInstances returns all available instance addresses.
func (p *ConnectionPool) GetAddresses() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	addresses := make([]string, 0, len(p.connections))
	for addr := range p.connections {
		addresses = append(addresses, addr)
	}
	return addresses
}

// Count returns the number of available connections.
func (p *ConnectionPool) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.connections)
}

// Close closes all connections in the pool.
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.cancel != nil {
		p.cancel()
	}

	var lastErr error
	for addr, conn := range p.connections {
		if conn != nil {
			if err := conn.Close(); err != nil && lastErr == nil {
				lastErr = fmt.Errorf("failed to close connection %s: %w", addr, err)
			}
		}
		delete(p.connections, addr)
	}

	return lastErr
}
