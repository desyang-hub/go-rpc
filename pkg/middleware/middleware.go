// Package middleware provides interceptor support for gRPC server and client.
//
// Middleware allows users to implement cross-cutting concerns such as
// logging, authentication, metrics collection, and rate limiting.
//
// The interceptor pattern follows the Unix pipeline model: each interceptor
// receives the request, optionally processes it, then passes it to the next
// interceptor in the chain or to the actual handler.
//
// # Server Setup
//
//	Middleware must be applied when creating the gRPC server. Use the
//	InterceptorChain helper to combine multiple interceptors into a single
//	unary or stream interceptor that will be passed to grpc.NewServer.
//
//	// Create individual interceptors
//	logging := NewLoggingInterceptor(logger)
//	auth := AuthInterceptor(tokenValidator)
//
//	// Chain them together
//	unaryChain := InterceptorChain(logging, auth)
//
//	// Create gRPC server with the chain
//	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(unaryChain))
//
// # Client Setup
//
//	Client interceptors are different — they use grpc.WithChainUnaryInterceptor
//	and must be passed as dial options.
//
package middleware

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// Interceptor is the interface for creating gRPC unary or stream interceptors.
// Implement this interface to create custom server-side middleware.
type Interceptor struct {
	unary   grpc.UnaryServerInterceptor
	stream  grpc.StreamServerInterceptor
}

func (i *Interceptor) Unary() grpc.UnaryServerInterceptor {
	return i.unary
}

func (i *Interceptor) Stream() grpc.StreamServerInterceptor {
	return i.stream
}

// NewInterceptor creates an interceptor from handler functions.
func NewInterceptor(unary grpc.UnaryServerInterceptor, stream grpc.StreamServerInterceptor) *Interceptor {
	if unary == nil {
		unary = defaultUnaryInterceptor
	}
	if stream == nil {
		stream = defaultStreamInterceptor
	}
	return &Interceptor{unary: unary, stream: stream}
}

// NewUnaryInterceptor creates an interceptor with only a unary handler.
func NewUnaryInterceptor(unary grpc.UnaryServerInterceptor) *Interceptor {
	return NewInterceptor(unary, nil)
}

// NewStreamInterceptor creates an interceptor with only a stream handler.
func NewStreamInterceptor(stream grpc.StreamServerInterceptor) *Interceptor {
	return NewInterceptor(nil, stream)
}

// defaultUnaryInterceptor is the passthrough interceptor that calls the next handler.
func defaultUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return handler(ctx, req)
}

// defaultStreamInterceptor is the passthrough interceptor that calls the next handler.
func defaultStreamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	return handler(srv, ss)
}

// InterceptorChain combines multiple interceptors into a single unary interceptor
// that executes them in order. This is the standard pattern for gRPC middleware.
func InterceptorChain(interceptors ...*Interceptor) grpc.UnaryServerInterceptor {
	var unwrapped []grpc.UnaryServerInterceptor
	for _, intercept := range interceptors {
		unwrapped = append(unwrapped, intercept.Unary())
	}
	return ChainUnaryServer(unwrapped...)
}

// ChainUnaryServer creates a unary interceptor that chains multiple interceptors.
func ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	var chain grpc.UnaryServerInterceptor
	chain = func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}

	for i := len(interceptors) - 1; i >= 0; i-- {
		i := i
		intercept := interceptors[i]
		prev := chain
		chain = func(next grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
			return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
				return intercept(ctx, req, info, func(innerCtx context.Context, innerReq interface{}) (interface{}, error) {
					if next == prev {
						return handler(innerCtx, innerReq)
					}
					return next(innerCtx, innerReq, info, handler)
				})
			}
		}(chain)
	}

	return chain
}

// Next returns the next handler in the interceptor chain.
func Next(unaryInterceptors []grpc.UnaryServerInterceptor, current int) grpc.UnaryServerInterceptor {
	n := len(unaryInterceptors)
	if current == n-1 {
		return nil
	}
	if current > n {
		return nil
	}
	return unaryInterceptors[current]
}

// normalizeMethod extracts the method name from a gRPC full method path.
func normalizeMethod(method string) string {
	if idx := strings.LastIndex(method, "/"); idx > 0 {
		return method[idx+1:]
	}
	return method
}

// normalizeService extracts the service name from a gRPC full method path.
func normalizeService(method string) string {
	if idx := strings.LastIndex(method, "/"); idx > 0 {
		return method[1:idx]
	}
	return method
}

// mapCodec is passed to the gRPC implementation for serialization.
type mapCodec struct{}

func (m mapCodec) NewDecoder(r interface{}) interface{} { return nil }
func (m mapCodec) NewEncoder(w interface{}) interface{} { return nil }

// RetryPolicy defines how failed RPC calls should be retried.
type RetryPolicy struct {
	MaxAttempts     int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	UsesExponential bool
	RetryableCodes  []int
}
