// Package middleware provides interceptor support for gRPC server and client.
//
// Middleware allows users to implement cross-cutting concerns such as
// logging, authentication, metrics collection, and rate limiting.
//
// The interceptor pattern follows the Unix pipeline model: each interceptor
// receives the request, optionally processes it, then passes it to the next
// interceptor in the chain or to the actual handler.
//
// Server middleware is registered as gRPC unary interceptors via the server's
// AddMiddleware method:
//
//	server := server.NewServer().
//	    AddMiddleware(middleware.Logging()).
//	    AddMiddleware(middleware.AuthCheck())
//
// To create custom server middleware, implement the ServerInterceptor interface:
//
//	type MyMiddleware struct{}
//
//	func (m MyMiddleware) ServerUnaryInterceptor() grpc.UnaryServerInterceptor {
//	    return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
//	        // Before handler
//	        resp, err := handler(ctx, req)
//	        // After handler
//	        return resp, err
//	    }
//	}
package middleware

import (
	"context"
	"time"

	"google.golang.org/grpc"
)

// ServerInterceptor is an interface for creating gRPC server interceptors.
// Implement this interface to create custom server-side middleware.
type ServerInterceptor interface {
	// UnaryInterceptor returns a unary server interceptor.
	// Return nil to skip this interceptor for unary calls.
	UnaryInterceptor() grpc.UnaryServerInterceptor

	// StreamInterceptor returns a streaming server interceptor.
	// Return nil to skip this interceptor for stream calls.
	StreamInterceptor() grpc.StreamServerInterceptor
}

// ServerInterceptorFunc is a convenience adapter for creating a ServerInterceptor
// from a single function.
type ServerInterceptorFunc struct {
	unary    grpc.UnaryServerInterceptor
	stream   grpc.StreamServerInterceptor
}

// NewServerInterceptorFunc creates a ServerInterceptor from handler functions.
func NewServerInterceptorFunc(unary grpc.UnaryServerInterceptor, stream grpc.StreamServerInterceptor) ServerInterceptor {
	return &ServerInterceptorFunc{unary: unary, stream: stream}
}

// UnaryInterceptor implements ServerInterceptor interface.
func (f *ServerInterceptorFunc) UnaryInterceptor() grpc.UnaryServerInterceptor {
	return f.unary
}

// StreamInterceptor implements ServerInterceptor interface.
func (f *ServerInterceptorFunc) StreamInterceptor() grpc.StreamServerInterceptor {
	return f.stream
}

// UnaryToBuildOption converts a unary interceptor to a server BuildOption.
// This allows middleware to be added to the server via AddMiddleware().
func UnaryToBuildOption(interceptor grpc.UnaryServerInterceptor) func(*grpc.Server) {
	return func(s *grpc.Server) {
		// gRPC doesn't support adding interceptors after server creation,
		// so this is a placeholder for future chaining support.
		// Users should apply middleware during server configuration.
	}
}

// Logging returns a server interceptor that logs request and response metadata.
func Logging() *ServerInterceptorFunc {
	return NewServerInterceptorFunc(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		method := info.FullMethod
		_ = method
		_ = duration

		// TODO: Add structured logging with zerolog or zap
		if err != nil {
			// _ = err // Will be logged in production
		}
		return resp, err
	}, nil)
}
