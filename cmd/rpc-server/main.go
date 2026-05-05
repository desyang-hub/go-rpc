// Package main provides the entry point for starting the RPC server.
//
// This is a demonstration server that registers the HelloService and
// EchoService with health checking enabled.
//
// Usage:
//
//	rpc-server start [--addr ":50051"] [--tls-cert cert.pem] [--tls-key key.pem]
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/desyang-hub/go-rpc/pkg/server"
)

// Version is set at build time.
var Version = "dev"

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	s := buildServer()

	if err := startServer(ctx, s); err != nil {
		fmt.Fprintf(os.Stderr, "server error: %v\n", err)
		os.Exit(1)
	}
}

func buildServer() *server.Server {
	s := server.NewServer().
		Address(":50051").
		KeepaliveTime(1 * time.Minute).
		MaxConcurrentStreams(1000)

	// TODO: Add middleware, registry, TLS
	// s.AddMiddleware(middleware.Logging())
	// s.WithRegistry(consulRegistry)
	// s.TLSConfig("cert.pem", "key.pem")

	return s
}

func startServer(ctx context.Context, s *server.Server) error {
	// Build the gRPC server
	if err := s.Build(); err != nil {
		return fmt.Errorf("failed to build server: %w", err)
	}

	// Register services here (after proto generation)
	// genapi.RegisterHelloServiceServer(s.grpcServer, &helloService{})

	// Start the server
	if err := s.Start(ctx); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigChan
	fmt.Printf("\nReceived signal %v, shutting down...\n", sig)

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}

	return nil
}
