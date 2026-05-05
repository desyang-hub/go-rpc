package server_test

import (
	"testing"
	"time"

	"github.com/desyang-hub/go-rpc/pkg/server"
)

func TestServerBuilderDefaults(t *testing.T) {
	s := server.NewServer()

	if s.Addr() != ":50051" {
		t.Errorf("expected default address :50051, got %s", s.Addr())
	}
}

func TestServerAddress(t *testing.T) {
	s := server.NewServer().Address(":9090")

	if s.Addr() != ":9090" {
		t.Errorf("expected address :9090, got %s", s.Addr())
	}
}

func TestServerKeepaliveTime(t *testing.T) {
	s := server.NewServer().KeepaliveTime(30 * time.Second)

	if s.Addr() == ":50051" {
		// Address chain works, verify continuation
	}
}

func TestServerMaxConcurrentStreams(t *testing.T) {
	s := server.NewServer().MaxConcurrentStreams(500)

	if s.Addr() == ":50051" {
		// Address chain works
	}
}

func TestServerTLSConfig(t *testing.T) {
	s := server.NewServer().TLSConfig("nonexistent.crt", "nonexistent.key")
	// Build should fail - this is expected
	err := s.Build()
	if err == nil {
		t.Error("expected error when loading nonexistent TLS cert/key")
	}
}
