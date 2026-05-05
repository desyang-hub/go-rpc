// Package middleware provides structured logging for gRPC requests and responses.
//
// The logging middleware captures request metadata, method details, execution
// time, and response errors. It uses zerolog for fast, structured JSON logging.
//
// # Usage
//
//	log := zerolog.New(os.Stdout).With().Timestamp().Logger()
//
//	logging := NewLoggingInterceptor(&LoggerImpl{Logger: log})
//	interceptor := logging.Unary()
//
package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// LoggerImpl is an interface for logging implementations.
// It allows injecting custom loggers (zerolog, zap, etc.).
type LoggerImpl interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	With(Tuple ...interface{}) *LoggerImpl
}

// zerologAdapter adapts zerolog.Logger to LoggerImpl interface.
type zerologAdapter struct {
	logger *zerolog.Logger
}

func (a *zerologAdapter) Debug(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		a.logger.Debug().Fields(fieldsToMap(fields...)).Msg(msg)
	} else {
		a.logger.Debug().Msg(msg)
	}
}

func (a *zerologAdapter) Info(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		a.logger.Info().Fields(fieldsToMap(fields...)).Msg(msg)
	} else {
		a.logger.Info().Msg(msg)
	}
}

func (a *zerologAdapter) Warn(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		a.logger.Warn().Fields(fieldsToMap(fields...)).Msg(msg)
	} else {
		a.logger.Warn().Msg(msg)
	}
}

func (a *zerologAdapter) Error(msg string, fields ...interface{}) {
	if len(fields) > 0 {
		a.logger.Error().Fields(fieldsToMap(fields...)).Msg(msg)
	} else {
		a.logger.Error().Msg(msg)
	}
}

func (a *zerologAdapter) With(Tuple ...interface{}) *LoggerImpl {
	return a
}

func fieldsToMap(fields ...interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			result[fmt.Sprintf("%v", fields[i])] = fields[i+1]
		}
	}
	return result
}

// LogFieldProvider extracts fields from gRPC metadata for logging.
type LogFieldProvider struct {
	IncludeMetadata []string
	ExcludeHeader   []string
}

func (p *LogFieldProvider) ShouldInclude(key string) bool {
	if len(p.ExcludeHeader) > 0 {
		for _, excluded := range p.ExcludeHeader {
			if key == excluded {
				return false
			}
		}
	}
	if len(p.IncludeMetadata) == 0 {
		return true
	}
	for _, included := range p.IncludeMetadata {
		if key == included {
			return true
		}
	}
	return false
}

// LogField provides logging fields from a gRPC request.
type LogField struct {
	Method    string
	Duration  time.Duration
	Status    string
	StartTime time.Time
	ClientIP  string
	UserAgent string
	RequestId string
	Metadata  map[string]string
}

// NewLoggingInterceptor returns a unary and stream server interceptor pair
// that logs request metadata, execution time, and response status.
func NewLoggingInterceptor(logger zerolog.Logger) *Interceptor {
	unary := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		// Extract request metadata
		var clientIP, userAgent, requestId string
		if p, ok := peer.FromContext(ctx); ok {
			if p.Addr != nil {
				clientIP = p.Addr.String()
			}
		}
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if ua := md.Get("user-agent"); len(ua) > 0 {
				userAgent = ua[0]
			}
			if rid := md.Get("x-request-id"); len(rid) > 0 {
				requestId = rid[0]
			}
		}

		// Call the actual handler
		resp, err := handler(ctx, req)

		// Log the result
		duration := time.Since(start)
		statusCode := "OK"
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code().String()
			} else {
				statusCode = "UNKNOWN"
			}
		}

		level := logger.Info()
		if statusCode != "OK" {
			level = logger.Error()
		}

		level.
			Str("method", normalizeMethod(info.FullMethod)).
			Str("service", normalizeService(info.FullMethod)).
			Str("client", clientIP).
			Str("request_id", requestId).
			Dur("duration_ms", duration).
			Str("status", statusCode).
			Msg("grpc_request")

		if err != nil {
			level = logger.Error()
		}

		return resp, err
	}

	stream := func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		// Extract metadata
		var clientIP, requestId string
		if p, ok := peer.FromContext(ss.Context()); ok {
			if p.Addr != nil {
				clientIP = p.Addr.String()
			}
		}
		if md, ok := metadata.FromIncomingContext(ss.Context()); ok {
			if rid := md.Get("x-request-id"); len(rid) > 0 {
				requestId = rid[0]
			}
		}

		// Call the actual stream handler
		err := handler(srv, ss)

		// Log the result
		duration := time.Since(start)
		statusCode := "OK"
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code().String()
			} else {
				statusCode = "UNKNOWN"
			}
		}

		level := logger.Info()
		if statusCode != "OK" {
			level = logger.Error()
		}

		level.
			Str("method", normalizeMethod(info.FullMethod)).
			Str("service", normalizeService(info.FullMethod)).
			Str("client", clientIP).
			Str("request_id", requestId).
			Dur("duration_ms", duration).
			Str("status", statusCode).
			Msg("grpc_stream")

		return err
	}

	return NewInterceptor(unary, stream)
}

// NewLoggingInterceptorFromPool returns a logging interceptor using the provided logger pointer.
func NewLoggingInterceptorFromPool(logger *zerolog.Logger) *Interceptor {
	return NewLoggingInterceptor(*logger)
}
