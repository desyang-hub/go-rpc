// Package client provides a high-level gRPC client builder with enterprise features.
//
// The Client allows configuring retry policies, circuit breakers,
// load balancers, and connection pooling in a fluent API style.
//
// Usage:
//
//	client := NewClient().
//	    Address("127.0.0.1:50051").
//	    RetryPolicy(RetryPolicy{
//	        MaxAttempts: 3,
//	        Backoff:    exponential,
//	    }).
//	    Build()
//
// After building, use client.Dial(ctx) to establish the connection.
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

	"github.com/desyang/go-rpc/internal/loadbalancer"
	"github.com/desyang/go-rpc/pkg/middleware"
)

// Config contains all client configuration.
type Config struct {
	// Network
	Address      string
	DialTimeout  time.Duration

	// TLS
	TLSConfig     *tls.Config
	Credentials   credentials.TransportCredentials

	// Keepalive
	KeepalivePeriod  time.Duration
	KeepaliveTimeout time.Duration

	// Retry
	MaxAttempts      int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	UsesExponential  bool
	RetryableCodes   []int // gRPC status codes to retry on

	// Service discovery
	RegistryAddress string
	RegistryType    string // "consul", "etcd"

	// Load balancing
	StrategyName string // "round_robin", "weighted_round_robin", "least_connection", "consistent_hash"
	LBKey        string // Key for consistent hashing

	// Custom dial options
	DialOptions []grpc.DialOption
}

// RetryPolicy defines how failed RPC calls should be retried.
type RetryPolicy struct {
	MaxAttempts     int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	UsesExponential bool
	RetryableCodes  []int
}

// DefaultRetryPolicy returns a retry policy with 3 attempts and exponential backoff.
func DefaultRetryPolicy() RetryPolicy {
	return RetryPolicy{
		MaxAttempts:     3,
		InitialBackoff:  100 * time.Millisecond,
		MaxBackoff:      3 * time.Second,
		UsesExponential: true,
		RetryableCodes:  []int{14, 4}, // UNAVAILABLE, DEADLINE_EXCEEDED
	}
}

// Client provides a configurable gRPC client with enterprise features.
type Client struct {
	config       Config
	conn         *grpc.ClientConn
	addr         string
	mu           sync.Mutex
	closed       chan struct{}
	lb           loadbalancer.Strategy
	registry     RegistryWatcher
	watchCancel  context.CancelFunc
}

// RegistryWatcher watches for service instance changes.
type RegistryWatcher interface {
	Watch(ctx context.Context) (<-chan ServiceUpdate, error)
	AddInstance(addr string)
	RemoveInstance(addr string)
}

// ServiceUpdate represents a change in service registration.
type ServiceUpdate struct {
	Service ServiceInfo
	Type    UpdateType
}

// ServiceInfo contains information about a registered service instance.
type ServiceInfo struct {
	Name     string
	Addr     string
	Metadata map[string]string
	TTL      time.Duration
	Healthy  bool
}

// UpdateType indicates the type of service change.
type UpdateType int

const (
	UpdateTypeRegister UpdateType = iota + 1
	UpdateTypeDeregister
	UpdateTypeUpdate
)

// NewClient creates a new Client with default settings.
func NewClient() *Client {
	cfg := defaultConfig()
	closed := make(chan struct{})
	c := &Client{
		config: cfg,
		closed: closed,
		addr:   cfg.Address,
	}
	return c
}

func defaultConfig() Config {
	return Config{
		Address:          "127.0.0.1:50051",
		DialTimeout:      10 * time.Second,
		KeepalivePeriod:  30 * time.Second,
		KeepaliveTimeout: 10 * time.Second,
		StrategyName:     "round_robin",
	}
}

// Address sets the server address to connect to.
func (c *Client) Address(addr string) *Client {
	c.addr = addr
	return c
}

// RetryPolicy sets the retry behavior for failed RPC calls.
func (c *Client) RetryPolicy(policy RetryPolicy) *Client {
	c.config.MaxAttempts = policy.MaxAttempts
	c.config.InitialBackoff = policy.InitialBackoff
	c.config.MaxBackoff = policy.MaxBackoff
	c.config.UsesExponential = policy.UsesExponential
	c.config.RetryableCodes = policy.RetryableCodes
	return c
}

// DialTimeout sets the timeout for establishing a connection.
func (c *Client) DialTimeout(d time.Duration) *Client {
	c.config.DialTimeout = d
	return c
}

// TLSConfig sets TLS configuration for secure connections.
func (c *Client) TLSConfig(config *tls.Config) *Client {
	c.config.TLSConfig = config
	return c
}

// TLSCredentials sets transport credentials directly.
func (c *Client) TLSCredentials(c credentials.TransportCredentials) *Client {
	c.config.Credentials = c
	return c
}

// KeepalivePeriod sets the keepalive ping period.
func (c *Client) KeepalivePeriod(d time.Duration) *Client {
	c.config.KeepalivePeriod = d
	return c
}

// KeepaliveTimeout sets the keepalive ping timeout.
func (c *Client) KeepaliveTimeout(d time.Duration) *Client {
	c.config.KeepaliveTimeout = d
	return c
}

// Middleware adds a gRPC interceptor chain to the client dial options.
func (c *Client) Middleware(chain *middleware.Interceptor) *Client {
	if chain != nil {
		c.config.DialOptions = append(c.config.DialOptions,
			grpc.WithChainUnaryInterceptor(chain.Unary()))
	}
	return c
}

// WithDialOptions adds custom grpc.DialOptions.
func (c *Client) WithDialOptions(opts ...grpc.DialOption) *Client {
	c.config.DialOptions = append(c.config.DialOptions, opts...)
	return c
}

// WithLoadBalancer sets the load balancer strategy.
func (c *Client) WithLoadBalancer(strategy string, key string) *Client {
	c.config.StrategyName = strategy
	c.config.LBKey = key
	c.lb = loadbalancer.NewStrategyFactory().NewStrategy(strategy)
	return c
}

// Dial dial the server and establish a connection.
func (c *Client) Dial(ctx context.Context) error {
	if err := c.Build(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		return nil // Already connected
	}

	opts := c.dialOptions()
	conn, err := grpc.DialContext(ctx, c.addr, opts...)
	if err != nil {
		return fmt.Errorf("failed to dial: %w", err)
	}

	c.conn = conn
	return nil
}

func (c *Client) dialOptions() []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(grpc.KeepaliveParams{
			Time:                c.config.KeepalivePeriod,
			Timeout:             c.config.KeepaliveTimeout,
			PermitWithoutStream: true,
		}),
		grpc.WithBlock(),
		grpc.WithDisableRetry(), // We implement our own retry
	}

	if c.config.DialTimeout > 0 {
		opts = append(opts, grpc.WithTimeout(c.config.DialTimeout))
	}

	if c.config.Credentials != nil {
		opts = append(opts, grpc.WithTransportCredentials(c.config.Credentials))
	} else if c.config.TLSConfig != nil {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(c.config.TLSConfig)))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	}

	return append(opts, c.config.DialOptions...)
}

// Build finalizes the client configuration.
func (c *Client) Build() error {
	// Validate configuration
	if c.addr == "" {
		return fmt.Errorf("address is required")
	}
	return nil
}

// Connection returns the underlying gRPC connection.
// Callers should NOT close this connection directly; use Client.Close() instead.
func (c *Client) Connection() *grpc.ClientConn {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn
}

// Close closes the client connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if s := c.conn; s != nil {
		return s.Close()
	}
	return nil
}

// IsClosed returns true if the client has been shut down.
func (c *Client) IsClosed() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn == nil {
		return true
	}
	// State 2 = connecting, 3 = ready, 4 = shutting down, 5 = error
	return c.conn.GetState() == 4 || c.conn.GetState() == 5
}
