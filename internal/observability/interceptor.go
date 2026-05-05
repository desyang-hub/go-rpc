package observability

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// ServerInterceptor returns a gRPC server interceptor that records metrics and tracing.
func ServerInterceptor(m *Metrics, t *Tracer, l *Logger, serviceName string) grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		spanCtx, span := t.StartSpan(ctx, info.FullMethod)
		_ = spanCtx

		release := m.ActiveConnections(serviceName)
		defer release()

		l.Debug("RPC call started", map[string]interface{}{
			"method": info.FullMethod,
			"request": req,
		})

		startTime := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(startTime)

		code := "unknown"
		if s, ok := status.FromError(err); ok {
			code = s.Code().String()
		}

		m.ReqFinished("server", serviceName, info.FullMethod, duration, code, err != nil)

		if span != nil {
			if err != nil {
				t.RecordSpanError(span, err)
			}
			t.RecordSpanDuration(span, duration)
		}

		entryMap := map[string]interface{}{
			"method":      info.FullMethod,
			"duration_us": duration.Microseconds(),
			"status":      code,
		}
		if err != nil {
			entryMap["error"] = err
			l.Error("RPC call failed", entryMap)
		} else {
			l.Debug("RPC call completed", entryMap)
		}

		if span != nil {
			span.End()
		}

		return resp, err
	})
}

// ClientInterceptor returns a gRPC client interceptor that records metrics and tracing.
func ClientInterceptor(m *Metrics, t *Tracer, l *Logger, serviceName string, serverAddr string) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, server any, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		attrs := t.SetClientSpanAttributes(serviceName, method, serverAddr)

		spanCtx, span := t.SpanWithAttributes(ctx, method, attrs...)
		_ = spanCtx

		startTime := time.Now()
		err := invoker(ctx, method, server, reply, cc, opts...)
		duration := time.Since(startTime)

		code := "unknown"
		if s, ok := status.FromError(err); ok {
			code = s.Code().String()
		}

		m.ReqFinished("client", serviceName, method, duration, code, err != nil)

		if span != nil {
			t.RecordSpanError(span, err)
			t.RecordSpanDuration(span, duration)
		}

		entryMap := map[string]interface{}{
			"method":      method,
			"duration_us": duration.Microseconds(),
			"status":      code,
		}
		if err != nil {
			entryMap["error"] = err
			l.Error("RPC call failed", entryMap)
		}

		if span != nil {
			span.End()
		}

		return err
	}
}
