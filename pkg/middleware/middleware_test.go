package middleware

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"google.golang.org/grpc"
)

func TestNewInterceptor(t *testing.T) {
	unary := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}

	intercept := NewInterceptor(unary, nil)
	if intercept == nil {
		t.Fatal("Expected non-nil interceptor")
	}
	if intercept.Unary() == nil {
		t.Error("Expected non-nil unary interceptor")
	}
}

func TestChainUnaryServer(t *testing.T) {
	var executed []string
	var mu sync.Mutex

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		mu.Lock()
		executed = append(executed, "handler")
		mu.Unlock()
		return req, nil
	}

	intercept1 := NewUnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (interface{}, error) {
		mu.Lock()
		executed = append(executed, "intercept1")
		mu.Unlock()
		return h(ctx, req)
	})

	intercept2 := NewUnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, h grpc.UnaryHandler) (interface{}, error) {
		mu.Lock()
		executed = append(executed, "intercept2")
		mu.Unlock()
		return h(ctx, req)
	})

	chain := InterceptorChain(intercept1, intercept2)

	ctx := context.Background()
	result, err := chain(ctx, "request", nil, handler)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if result != "request" {
		t.Fatalf("Expected 'request', got %v", result)
	}

	mu.Lock()
	if len(executed) != 3 {
		t.Errorf("Expected 3 executions, got %d: %v", len(executed), executed)
	}
	mu.Unlock()
}

func TestNormalizeMethod(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"/package.Service/Method", "Method"},
		{"/Service/Method", "Method"},
		{"Method", "Method"},
	}

	for _, tt := range tests {
		got := normalizeMethod(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeMethod(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNormalizeService(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"/package.Service/Method", "package.Service"},
		{"/Service/Method", "Service"},
		{"Service/Method", "Service"},
	}

	for _, tt := range tests {
		got := normalizeService(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeService(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestNewRequestInfo(t *testing.T) {
	ri := NewRequestInfo("/package.Service/Method")

	if ri.FullMethod != "/package.Service/Method" {
		t.Errorf("FullMethod = %q, want %q", ri.FullMethod, "/package.Service/Method")
	}

	if ri.Method != "Method" {
		t.Errorf("Method = %q, want %q", ri.Method, "Method")
	}

	if ri.Service != "package.Service" {
		t.Errorf("Service = %q, want %q", ri.Service, "package.Service")
	}
}

func TestContainsStatusCode(t *testing.T) {
	codeList := []int{14, 4} // UNAVAILABLE, DEADLINE_EXCEEDED

	if containsStatusCode(14, codeList) != true {
		t.Error("Expected true for code 14 in list")
	}

	if containsStatusCode(0, codeList) != false {
		t.Error("Expected false for code 0 not in list")
	}
}
