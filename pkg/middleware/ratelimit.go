// Package middleware provides rate limiting as a gRPC interceptor.
//
// The rate limiting middleware uses a token bucket algorithm to control
// the rate of requests. It can be configured globally or per-client/IP.
//
// # Usage
//
//	limit, _ := NewRateLimiter(&RateLimitConfig{
//	    RequestsPerSecond: 100,
//	    BurstSize:         200,
//	    PerClient:         true,
//	})
//
//	interceptor := limit.Interceptor()
//
// During server creation:
//
//	grpc.NewServer(grpc.UnaryInterceptor(middleware.InterceptorChain(interceptor)))
package middleware

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// RateLimiter implements a token bucket rate limiting algorithm.
// Each client can have its own bucket when PerClient mode is enabled.
type RateLimiter struct {
	mux          sync.RWMutex
	buckets      map[string]*bucket
	config       RateLimitConfig
	cleanupTicker *time.Ticker
	cleanupStop  chan struct{}
	pending      chan struct{}
}

// bucket represents a single token bucket for rate limiting.
type bucket struct {
	tokens     float64
	refillRate float64
	maxTokens  float64
	lastUpdate time.Time
	mu         sync.Mutex
}

// RateLimitConfig contains rate limiter configuration.
type RateLimitConfig struct {
	// RequestsPerSecond controls the maximum sustained request rate
	RequestsPerSecond float64
	// BurstSize allows temporary spikes above the sustained rate
	BurstSize int
	// PerClient enables per-client rate limiting by IP
	PerClient bool
	// PathFilter allows limiting only specific RPC paths
	PathFilter *PathFilter
	// TTL controls how long an idle client bucket is kept
	TTL time.Duration
}

// PathFilter allows rate limiting to be applied only to specific RPC paths.
type PathFilter struct {
	Include []string
	Exclude []string
}

// fullPathMatches checks if a gRPC method path matches the filter.
func (f *PathFilter) fullPathMatches(path string) bool {
	if f == nil {
		return true
	}

	// If include is specified, path must be in it
	if len(f.Include) > 0 {
		found := false
		for _, p := range f.Include {
			if p == path {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	// If exclude is specified, path must not be in it
	for _, p := range f.Exclude {
		if p == path {
			return false
		}
	}

	return true
}

// NewRateLimiter creates a new RateLimiter with the given configuration.
// It starts a background cleanup goroutine.
func NewRateLimiter(cfg *RateLimitConfig) (*RateLimiter, error) {
	if cfg == nil {
		cfg = DefaultRateLimitConfig()
	}

	if cfg.RequestsPerSecond <= 0 {
		return nil, fmt.Errorf("requests per second must be positive")
	}
	if cfg.BurstSize <= 0 {
		return nil, fmt.Errorf("burst size must be positive")
	}

	ttl := cfg.TTL
	if ttl == 0 {
		ttl = 10 * time.Minute
	}

	rl := &RateLimiter{
		buckets:      make(map[string]*bucket),
		config:       *cfg,
		cleanupStop:  make(chan struct{}),
	}

	rl.pending = make(chan struct{}, cfg.BurstSize)

	// Start cleanup goroutine
	rl.cleanupTicker = time.NewTicker(5 * time.Minute)
	go rl.cleanupLoop()

	return rl, nil
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := rl.cleanupTicker
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupIdle()
		case <-rl.cleanupStop:
			return
		}
	}
}

func (rl *RateLimiter) cleanupIdle() {
	rl.mux.Lock()
	defer rl.mux.Unlock()

	network := rl.config.PerClient
	deadline := time.Now().Add(-rl.config.TTL)

	for id, bkt := range rl.buckets {
		bkt.mu.Lock()
		if network && bkt.lastUpdate.Before(deadline) {
			bkt.mu.Unlock()
			delete(rl.buckets, id)
		} else if !network && bkt.lastUpdate.Before(deadline) {
			bkt.mu.Unlock()
			delete(rl.buckets, id)
		} else {
			bkt.mu.Unlock()
		}
	}
}

// keyForClient returns the key to identify a client for rate limiting.
func (rl *RateLimiter) keyForClient(ctx context.Context) string {
	if rl.config.PerClient {
		p, ok := peer.FromContext(ctx)
		if ok && p != nil && p.Addr != nil {
			return p.Addr.String()
		}
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if clientID := md.Get("x-client-id"); len(clientID) > 0 {
			return "client:" + clientID[0]
		}
		if ip := md.Get("x-real-ip"); len(ip) > 0 {
			return "ip:" + ip[0]
		}
	}

	return "global"
}

// allow checks if a request is allowed under the rate limit.
func (rl *RateLimiter) allow(ctx context.Context, method string) bool {
	// Check path filter
	if rl.config.PathFilter != nil && !rl.config.PathFilter.fullPathMatches(method) {
		return true
	}

	key := rl.keyForClient(ctx)

	rl.mux.Lock()
	bkt, exists := rl.buckets[key]
	rl.mux.Unlock()

	if !exists {
		bkt = &bucket{
			tokens:     float64(rl.config.BurstSize),
			refillRate: rl.config.RequestsPerSecond,
			maxTokens:  float64(rl.config.BurstSize),
		}
		rl.mux.Lock()
		rl.buckets[key] = bkt
		rl.mux.Unlock()
	}

	bkt.mu.Lock()
	defer bkt.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(bkt.lastUpdate).Seconds()
	bkt.lastUpdate = now

	bkt.tokens += elapsed * bkt.refillRate
	if bkt.tokens > bkt.maxTokens {
		bkt.tokens = bkt.maxTokens
	}

	if bkt.tokens >= 1.0 {
		bkt.tokens -= 1.0
		return true
	}

	return false
}

// report returns the reported rate limit error
func (rl *RateLimiter) report(ctx context.Context, method string) {
	status.New(codes.ResourceExhausted,
		fmt.Sprintf("rate limit exceeded: method=%s (limit=%.0f req/s, burst=%d)",
			method, rl.config.RequestsPerSecond, rl.config.BurstSize)).Send()
}

// Interceptor returns a unary server interceptor for rate limiting.
func (rl *RateLimiter) Interceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if !rl.allow(ctx, info.FullMethod) {
			return nil, status.Error(codes.ResourceExhausted,
				fmt.Sprintf("rate limit exceeded: %s (limit=%.0f req/s, burst=%d)",
					info.FullMethod, rl.config.RequestsPerSecond, rl.config.BurstSize))
		}
		return handler(ctx, req)
	}
}

// Stop stops the rate limiter cleanup loop.
func (rl *RateLimiter) Stop() {
	close(rl.cleanupStop)
	if rl.cleanupTicker != nil {
		rl.cleanupTicker.Stop()
	}
}

// Stats returns the current rate limit stats.
func (rl *RateLimiter) Stats() map[string]float64 {
	stats := make(map[string]float64)
	rl.mux.Lock()
	defer rl.mux.Unlock()

	for id, bkt := range rl.buckets {
		bkt.mu.Lock()
		stats[id] = math.Round(bkt.tokens*100) / 100
		bkt.mu.Unlock()
	}

	return stats
}

// DefaultRateLimitConfig returns a reasonable default configuration.
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerSecond: 100,
		BurstSize:         200,
		PerClient:         true,
		TTL:               10 * time.Minute,
	}
}
