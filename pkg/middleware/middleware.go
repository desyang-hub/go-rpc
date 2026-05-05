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
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

// Interceptor is the interface for creating gRPC unary or stream interceptors.
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
// that executes them in order.
func InterceptorChain(interceptors ...*Interceptor) grpc.UnaryServerInterceptor {
	var unwrapped []grpc.UnaryServerInterceptor
	for _, intercept := range interceptors {
		unwrapped = append(unwrapped, intercept.Unary())
	}
	return ChainUnaryServer(unwrapped...)
}

// ChainUnaryServer creates a unary interceptor chain.
func ChainUnaryServer(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return callChain(ctx, req, info, handler, interceptors, 0)
	}
}

// callChain recursively calls each interceptor.
func callChain(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler, interceptors []grpc.UnaryServerInterceptor, index int) (interface{}, error) {
	if index >= len(interceptors) {
		return handler(ctx, req)
	}
	current := interceptors[index]
	return current(ctx, req, info, func(innerCtx context.Context, innerReq interface{}) (interface{}, error) {
		return callChain(innerCtx, innerReq, info, handler, interceptors, index+1)
	})
}

// ExtractClientIP extracts the client IP address from the gRPC context.
func ExtractClientIP(ctx context.Context) string {
	if p, ok := peer.FromContext(ctx); ok {
		if p != nil && p.Addr != nil {
			return p.Addr.String()
		}
	}

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ip := md.Get("x-real-ip"); len(ip) > 0 {
			return ip[0]
		}
		if ip := md.Get("x-forwarded-for"); len(ip) > 0 {
			return ip[0]
		}
		if ip := md.Get("x-client-ip"); len(ip) > 0 {
			return ip[0]
		}
	}

	return "unknown"
}

// ExtractHeader extracts a header value from the gRPC context.
func ExtractHeader(ctx context.Context, key string) string {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if vals := md.Get(key); len(vals) > 0 {
			return vals[0]
		}
	}
	return ""
}

// WithRequestID stores a request ID in context metadata
func WithRequestID(ctx context.Context, id string) context.Context {
	md := metadata.New(map[string]string{"x-request-id": id})
	return metadata.NewOutgoingContext(ctx, md)
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

// RetryPolicy defines how failed RPC calls should be retried.
type RetryPolicy struct {
	MaxAttempts     int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	UsesExponential bool
	RetryableCodes  []int
}

// RequestInfo holds metadata from a gRPC request
type RequestInfo struct {
	FullMethod string
	Method     string
	Service    string
}

// NewRequestInfo creates a RequestInfo from a gRPC method path
func NewRequestInfo(fullMethod string) *RequestInfo {
	return &RequestInfo{
		FullMethod: fullMethod,
		Method:     normalizeMethod(fullMethod),
		Service:    normalizeService(fullMethod),
	}
}

// containsStatusCode checks if the given status code is in the list
func containsStatusCode(code codes.Code, codes []int) bool {
	for _, c := range codes {
		if int(code) == c {
			return true
		}
	}
	return false
}

// getTokenFromContext extracts the token from Authorization header
func getTokenFromContext(ctx context.Context) string {
	auth := ExtractHeader(ctx, "authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:]
	}
	return ""
}

// recordRequest logs request start time for metrics
func recordRequest(ctx context.Context) context.Context {
	return context.WithValue(ctx, "request_start", time.Now())
}

// requestDuration calculates duration from context
func requestDuration(ctx context.Context) time.Duration {
	if start, ok := ctx.Value("request_start").(time.Time); ok {
		return time.Since(start)
	}
	return 0
}
