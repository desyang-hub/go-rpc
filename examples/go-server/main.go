// Package main implements the HelloService example server.
// This demonstrates a production-ready RPC server with:
// - Four gRPC call modes (unary, server-stream, client-stream, bidirectional-stream)
// - Prometheus metrics endpoint
// - Health check endpoint
// - Graceful shutdown
// - Structured logging with slog
package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"


	"go-rpc/pkg/server"
	pb "go-rpc/api/genapi"
	"go-rpc/internal/observability"
	"go-rpc/internal/healthcheck"
)

const (
	ServerAddr  = ":50051"
	MetricsAddr = ":9090"
	HealthAddr  = ":8081"
	ServerID    = "go-rpc-server-1"
)

// Service0 is our HelloService implementation providing all four RPC call modes.
type Service0 struct {
	mu      sync.Mutex
	start   time.Time
	metrics *observability.Metrics
}

// NewService0 creates a new service implementation.
func NewService0() *Service0 {
	return &Service0{
		start:   time.Now(),
		metrics: observability.NewMetrics(),
	}
}

// Hello implements a unary RPC (request-response).
func (s *Service0) Hello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	t0 := time.Now()
	uptime := time.Since(s.start)

	resp := &pb.HelloResponse{
		Message:   fmt.Sprintf("Hello, %s! Server up %v.", req.GetName(), uptime.Truncate(time.Second)),
		Timestamp: time.Now().Format(time.RFC3339),
		ServerId:  ServerID,
	}

	s.metrics.ReqFinished("server", "HelloService", "hello", time.Since(t0), "", false)
	slog.Info("Hello RPC", "name", req.GetName(), "mode", "unary")
	return resp, nil
}

// HelloStream implements a server-streaming RPC (1→N).
func (s *Service0) HelloStream(req *pb.HelloRequest, srv pb.HelloService_HelloStreamServer) error {
	t0 := time.Now()
	name := req.GetName()

	for i := 0; i < 5; i++ {
		select {
		case <-srv.Context().Done():
			return srv.Context().Err()
		default:
		}
		if err := srv.Send(&pb.HelloStreamResponse{
			Message:       fmt.Sprintf("Stream %d for %s", i+1, name),
			ResponseIndex: int32(i + 1),
		}); err != nil {
			return err
		}
		time.Sleep(100 * time.Millisecond)
	}
	s.metrics.ReqFinished("server", "HelloService", "hello_stream", time.Since(t0), "", false)
	slog.Info("HelloStream completed", "name", name)
	return nil
}

// BatchHello implements a client-streaming RPC (N→1).
func (s *Service0) BatchHello(srv pb.HelloService_BatchHelloServer) error {
	t0 := time.Now()
	var names []string
	for {
		req, err := srv.Recv()
		if err == io.EOF { break }
		if err != nil { return err }
		names = append(names, req.GetName())
	}
	summ := fmt.Sprintf("%d users", len(names))
	if len(names) == 1 { summ = names[0] }
	resp := &pb.HelloResponse{
		Message:   fmt.Sprintf("Batch: %s (%d reqs)", summ, len(names)),
		Timestamp: time.Now().Format(time.RFC3339),
		ServerId:  ServerID,
	}
	if err := srv.SendAndClose(resp); err != nil { return err }
	s.metrics.ReqFinished("server", "HelloService", "batch_hello", time.Since(t0), "", false)
	slog.Info("BatchHello completed", "count", len(names))
	return nil
}

// HelloStreamStream implements a bidirectional streaming RPC (N⇄N).
func (s *Service0) HelloStreamStream(srv pb.HelloService_HelloStreamStreamServer) error {
	t0 := time.Now()
	i := 0
	for {
		req, err := srv.Recv()
		if err == io.EOF { break }
		if err != nil { return err }
		i++
		if err := srv.Send(&pb.HelloStreamResponse{
			Message:       fmt.Sprintf("Echo %d: %s", i, req.GetName()),
			ResponseIndex: int32(i),
		}); err != nil { return err }
	}
	s.metrics.ReqFinished("server", "HelloService", "hello_stream_stream", time.Since(t0), "", false)
	slog.Info("HelloStreamStream completed", "requests", i)
	return nil
}

// healthHandler implements /health HTTP endpoint.
type healthHandler struct{ start time.Time }

func (h *healthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.URL.Path != "/health" { http.NotFound(w, r); return }
	body := fmt.Sprintf(`{"status": "healthy", "uptime": "%v"}`,
		time.Since(h.start).Truncate(time.Second))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(body))
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Observability
	op := observability.NewMetrics()
	hc := healthcheck.New(nil, healthcheck.DefaultConfig())

	// HTTP endpoints
	mux := http.NewServeMux()
	mux.Handle("/health", &healthHandler{start: time.Now()})
	go http.ListenAndServe(HealthAddr, mux)

	muxM := http.NewServeMux()
	muxM.Handle("/metrics", op.HTTPHandler())
	go http.ListenAndServe(MetricsAddr, muxM)

	// gRPC server
	svc := NewService0()
	s := server.NewServer().Address(ServerAddr).KeepaliveTime(30 * time.Second).MaxConcurrentStreams(100)
	if err := s.Build(); err != nil { slog.Error("build", "err", err); os.Exit(1) }
	pb.RegisterHelloServiceServer(s.GRPCServer(), svc)

	slog.Info("Starting server", "grpc", ServerAddr, "metrics", MetricsAddr, "health", HealthAddr)
	if err := s.Start(ctx); err != nil && err != context.Canceled { slog.Error("start", "err", err) }

	<-ctx.Done()
	slog.Info("Shutdown")
	cancel()
	_ = s.Shutdown(ctx)
	slog.Info("Stopped")
}
