// Package server provides a high-level gRPC server builder with enterprise features.
//
// The Server allows configuring TLS, keepalive, connection limits,
// middleware chains, and service registration in a fluent API style.
//
// Usage:
//
//	server := NewServer().
//	    Address(":50051").
//	    KeepaliveTime(30 * time.Second).
//	    AddMiddleware(middleware.LoggerInterceptor()).
//	    WithRegistry(registry).
//	    Build() // Build the grpc.Server internally
//
// If you need to register additional services, use server.RegisterService().
package server

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpckeepalive "google.golang.org/grpc/keepalive"
)

// BuildOption is a functional option for configuring the gRPC server.
// It applies post-build modifications (e.g. middleware registration).
type BuildOption func(*grpc.Server)

// ServiceInfo contains information about a registered service instance.
type ServiceInfo struct {
	Name     string
	Addr     string
	Metadata map[string]string
	TTL      time.Duration
	Healthy  bool
}

// RegistryBuilder creates a new registry instance for service discovery.
type RegistryBuilder interface {
	Build() RegistryInstance
}

// RegistryInstance is the interface for service registration and discovery.
type RegistryInstance interface {
	Register(ctx context.Context, service ServiceInfo) error
	Deregister(ctx context.Context, service ServiceInfo) error
	Watch(ctx context.Context) (<-chan ServiceInfo, error)
}

// Config contains all server configuration.
type Config struct {
	Address          string
	EnableTLS        bool
	CertFile         string
	KeyFile          string
	TLSConfig        *tls.Config
	Credentials      credentials.TransportCredentials
	KeepaliveTime    time.Duration
	KeepaliveTimeout time.Duration
	MaxConcurrentStreams uint32
	MaxSendMsgSize   int
	MaxRecvMsgSize   int
	MaxConnectionIdle    time.Duration
	MaxConnectionAge     time.Duration
	MaxConnectionAgeGrace time.Duration
	RegistryBuilder      RegistryBuilder
	BuildOptions         []BuildOption
}

// Server provides a configurable gRPC server with enterprise features.
type Server struct {
	config     Config
	grpcServer *grpc.Server
	addr       string
	listener   net.Listener
	wg         sync.WaitGroup
	closed     chan struct{}
	registry   RegistryInstance
}

// NewServer creates a new Server with default settings.
func NewServer() *Server {
	cfg := defaultConfig()
	closed := make(chan struct{})
	s := &Server{
		config: cfg,
		closed: closed,
		addr:   cfg.Address,
	}
	return s
}

func defaultConfig() Config {
	return Config{
		Address:              ":50051",
		KeepaliveTime:        5 * time.Minute,
		KeepaliveTimeout:     20 * time.Second,
		MaxConcurrentStreams: 100,
	}
}

// HasKeyPair returns true if both cert and key files are set.
func (c *Config) HasKeyPair() bool {
	return c.CertFile != "" && c.KeyFile != ""
}

// Address sets the server listen address.
func (s *Server) Address(addr string) *Server {
	s.addr = addr
	return s
}

// TLSConfig sets TLS configuration for secure connections.
// Both certFile and keyFile must be provided to enable TLS.
func (s *Server) TLSConfig(certFile, keyFile string) *Server {
	s.config.CertFile = certFile
	s.config.KeyFile = keyFile
	s.config.EnableTLS = true
	return s
}

// TLSCredentials sets transport credentials directly.
func (s *Server) TLSCredentials(c credentials.TransportCredentials) *Server {
	s.config.Credentials = c
	return s
}

// KeepaliveTime sets the keepalive ping interval.
func (s *Server) KeepaliveTime(d time.Duration) *Server {
	s.config.KeepaliveTime = d
	return s
}

// KeepaliveTimeout sets the keepalive ping timeout.
func (s *Server) KeepaliveTimeout(d time.Duration) *Server {
	s.config.KeepaliveTimeout = d
	return s
}

// MaxConcurrentStreams sets the maximum number of concurrent streams per connection.
func (s *Server) MaxConcurrentStreams(n uint32) *Server {
	s.config.MaxConcurrentStreams = n
	return s
}

// MaxSendMsgSize sets the maximum message size allowed for send (downstream).
func (s *Server) MaxSendMsgSize(n int) *Server {
	s.config.MaxSendMsgSize = n
	return s
}

// MaxRecvMsgSize sets the maximum message size allowed for receive (upstream).
func (s *Server) MaxRecvMsgSize(n int) *Server {
	s.config.MaxRecvMsgSize = n
	return s
}

// MaxConnectionIdle sets the maximum duration a connection can remain idle.
func (s *Server) MaxConnectionIdle(d time.Duration) *Server {
	s.config.MaxConnectionIdle = d
	return s
}

// MaxConnectionAge sets the maximum duration a connection can exist.
func (s *Server) MaxConnectionAge(d time.Duration) *Server {
	s.config.MaxConnectionAge = d
	return s
}

// MaxConnectionAgeGrace sets additional time after MaxConnectionAge before closing.
func (s *Server) MaxConnectionAgeGrace(d time.Duration) *Server {
	s.config.MaxConnectionAgeGrace = d
	return s
}

// WithRegistry sets up service registration and discovery.
func (s *Server) WithRegistry(rb RegistryBuilder) *Server {
	s.config.RegistryBuilder = rb
	return s
}

// WithBuildOptions adds custom grpc.Server build options.
func (s *Server) WithBuildOptions(opts ...BuildOption) *Server {
	s.config.BuildOptions = append(s.config.BuildOptions, opts...)
	return s
}

// AddMiddleware adds a middleware (as BuildOption) to the server.
func (s *Server) AddMiddleware(opts ...BuildOption) *Server {
	s.config.BuildOptions = append(s.config.BuildOptions, opts...)
	return s
}

// Build finalizes the server configuration and creates the gRPC server.
// This must be called before Start().
func (s *Server) Build() error {
	opts := []grpc.ServerOption{
		grpc.MaxConcurrentStreams(s.config.MaxConcurrentStreams),
		grpc.MaxSendMsgSize(s.config.MaxSendMsgSize),
		grpc.MaxRecvMsgSize(s.config.MaxRecvMsgSize),
		grpc.KeepaliveParams(grpckeepalive.ServerParameters{
			Time:    s.config.KeepaliveTime,
			Timeout: s.config.KeepaliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(grpckeepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}


	if s.config.MaxConnectionIdle > 0 || s.config.MaxConnectionAge > 0 || s.config.MaxConnectionAgeGrace > 0 {
		params := grpckeepalive.ServerParameters{}
		if s.config.MaxConnectionIdle > 0 {
			params.MaxConnectionIdle = s.config.MaxConnectionIdle
		}
		if s.config.MaxConnectionAge > 0 {
			params.MaxConnectionAge = s.config.MaxConnectionAge
		}
		if s.config.MaxConnectionAgeGrace > 0 {
			params.MaxConnectionAgeGrace = s.config.MaxConnectionAgeGrace
		}
		opts = append(opts, grpc.KeepaliveParams(params))
	}

	// Handle TLS via credentials
	if s.config.Credentials != nil {
		opts = append(opts, grpc.Creds(s.config.Credentials))
	} else if s.config.EnableTLS {
		// Load cert/key pair
		if !s.config.HasKeyPair() {
			return fmt.Errorf("TLS enabled but no cert/key provided")
		}
		cert, err := tls.LoadX509KeyPair(s.config.CertFile, s.config.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS cert/key: %w", err)
		}
		opts = append(opts, grpc.Creds(credentials.NewTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		})))
	}

	s.grpcServer = grpc.NewServer(opts...)

	// Apply build options (post-build callbacks, e.g. middleware registration).
	for _, m := range s.config.BuildOptions {
		if m != nil {
			m(s.grpcServer)
		}
	}
	return nil
}

// Start starts the gRPC server and registers services.
func (s *Server) Start(ctx context.Context) error {
	if s.grpcServer == nil {
		return fmt.Errorf("server not configured: call Build() first")
	}

	addr := s.addr
	if addr == "" {
		addr = ":50051"
	}

	lc := &net.ListenConfig{}
	listener, err := lc.Listen(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.listener = listener

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.grpcServer.Serve(s.listener)
	}()

	// Register with service discovery if configured
	if s.config.RegistryBuilder != nil {
		s.startRegistry(ctx)
	}

	fmt.Printf("Server is listening on %s\n", addr)
	return nil
}

// startRegistry registers the server with the service registry.
func (s *Server) startRegistry(ctx context.Context) {
	if s.config.RegistryBuilder == nil {
		return
	}

	_, port, err := net.SplitHostPort(s.addr)
	if err != nil {
		port = "50051"
	}

	s.registry = s.config.RegistryBuilder.Build()
	info := ServiceInfo{
		Name:     "go-rpc.server",
		Addr:     net.JoinHostPort("0.0.0.0", port),
		Metadata: map[string]string{"protocol": "gRPC"},
		Healthy:  true,
	}

	go func() {
		<-ctx.Done()
		if err := s.registry.Deregister(ctx, info); err != nil {
			fmt.Printf("warning: failed to deregister service: %v\n", err)
		}
	}()

	if err := s.registry.Register(ctx, info); err != nil {
		fmt.Printf("warning: failed to register service: %v\n", err)
	}
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	if s.grpcServer == nil {
		return nil
	}

	// Deregister from registry
	if s.registry != nil {
		addr := s.addr
		if addr == "" {
			addr = ":50051"
		}
		_, port, _ := net.SplitHostPort(addr)
		if port == "" {
			port = "50051"
		}
		info := ServiceInfo{
			Name:    "go-rpc.server",
			Addr:    net.JoinHostPort("0.0.0.0", port),
			Healthy: false,
		}
		_ = s.registry.Deregister(ctx, info)
	}

	s.grpcServer.GracefulStop()
	s.wg.Wait()
	close(s.closed)
	return nil
}

// IsClosed returns true if the server has been shut down.
func (s *Server) IsClosed() bool {
	select {
	case <-s.closed:
		return true
	default:
		return false
	}
}

// Listener returns the underlying network listener.
func (s *Server) Listener() net.Listener {
	return s.listener
}

// Close closes the server listener.
func (s *Server) Close() error {
	if s.listener != nil {
		return s.listener.Close()
	}
	return nil
}

// Addr returns the server listen address.
func (s *Server) Addr() string {
	return s.addr
}


// GRPCServer returns the underlying grpc.Server instance for service registration.
func (s *Server) GRPCServer() *grpc.Server {
	return s.grpcServer
}
