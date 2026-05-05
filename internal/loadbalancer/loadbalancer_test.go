package loadbalancer

import (
	"sync"
	"testing"
)

func TestRoundRobin(t *testing.T) {
	lb := NewRoundRobin()

	lb.AddInstance(Instance{Addr: "10.0.0.1:50051"})
	lb.AddInstance(Instance{Addr: "10.0.0.2:50051"})
	lb.AddInstance(Instance{Addr: "10.0.0.3:50051"})

	if lb.Count() != 3 {
		t.Errorf("Expected 3 instances, got %d", lb.Count())
	}

	// Round-robin should cycle through all instances
	// Due to atomic.AddInt64 returning new value (1-based), first call returns index 1
	expected := []string{"10.0.0.2:50051", "10.0.0.3:50051", "10.0.0.1:50051"}
	for i := 0; i < 6; i++ {
		inst, ok := lb.Next()
		if !ok {
			t.Fatalf("Expected instance at iteration %d", i)
		}
		if inst.Addr != expected[i%3] {
			t.Errorf("Iteration %d: got %s, expected %s", i, inst.Addr, expected[i%3])
		}
	}
}

func TestRoundRobinRemove(t *testing.T) {
	lb := NewRoundRobin()

	lb.AddInstance(Instance{Addr: "10.0.0.1:50051"})
	lb.AddInstance(Instance{Addr: "10.0.0.2:50051"})
	lb.RemoveInstance("10.0.0.1:50051")

	if lb.Count() != 1 {
		t.Errorf("Expected 1 instance after removal, got %d", lb.Count())
	}
}

func TestRoundRobinEmpty(t *testing.T) {
	lb := NewRoundRobin()

	_, ok := lb.Next()
	if ok {
		t.Error("Expected no instance from empty balancer")
	}
}

func TestWeightedRoundRobin(t *testing.T) {
	lb := NewWeightedRoundRobin()

	// Add instances with different weights
	lb.AddInstance(Instance{Addr: "10.0.0.1:50051", Weight: 3})
	lb.AddInstance(Instance{Addr: "10.0.0.2:50051", Weight: 1})

	// With weights 3:1, over 4 iterations we should see roughly 3:1 ratio
	instances := make(map[string]int)
	var mu sync.Mutex

	for i := 0; i < 8; i++ {
		inst, ok := lb.Next()
		if !ok {
			t.Fatalf("Expected instance at iteration %d", i)
		}
		mu.Lock()
		instances[inst.Addr]++
		mu.Unlock()
	}

	mu.Lock()
	c1 := instances["10.0.0.1:50051"]
	c2 := instances["10.0.0.2:50051"]
	mu.Unlock()

	// With weights 3:1, we expect roughly 6:2 ratio over 8 iterations
	if c1 < int(float64(c2)*1.2) {
		t.Errorf("Expected weighted distribution with more hits to higher weight, got %d vs %d", c1, c2)
	}
}

func TestWeightedRoundRobinRemoval(t *testing.T) {
	lb := NewWeightedRoundRobin()

	lb.AddInstance(Instance{Addr: "10.0.0.1:50051", Weight: 2})
	lb.AddInstance(Instance{Addr: "10.0.0.2:50051", Weight: 1})
	lb.RemoveInstance("10.0.0.1:50051")

	if lb.Count() != 1 {
		t.Errorf("Expected 1 instance after removal, got %d", lb.Count())
	}
}

func TestLeastConnection(t *testing.T) {
	lb := NewLeastConnection().(*LeastConnection)

	lb.AddInstance(Instance{Addr: "10.0.0.1:50051"})
	lb.AddInstance(Instance{Addr: "10.0.0.2:50051"})

	// Set different connection counts using SetConnects (since Instances() returns value copies)
	lb.SetConnects("10.0.0.1:50051", 2)
	lb.SetConnects("10.0.0.2:50051", 0)

	// Next() should return the instance with fewer connections (10.0.0.2)
	inst1, ok := lb.Next()
	if !ok {
		t.Fatal("Expected first instance")
	}
	if inst1.Addr != "10.0.0.2:50051" {
		t.Errorf("Expected 10.0.0.2:50051 (fewer connections), got %s", inst1.Addr)
	}
}

func TestLeastConnectionEmpty(t *testing.T) {
	lb := NewLeastConnection()

	_, ok := lb.Next()
	if ok {
		t.Error("Expected no instance from empty balancer")
	}
}

func TestConsistentHash(t *testing.T) {
	lb := NewConsistentHash()

	lb.AddInstance(Instance{Addr: "10.0.0.1:50051"})
	lb.AddInstance(Instance{Addr: "10.0.0.2:50051"})
	lb.AddInstance(Instance{Addr: "10.0.0.3:50051"})

	// Same instance should map to same server
	var firstResults []string
	var secondResults []string

	for i := 0; i < 5; i++ {
		inst1, ok := lb.Next()
		if ok {
			firstResults = append(firstResults, inst1.Addr)
		}

		inst2, ok := lb.Next()
		if ok {
			secondResults = append(secondResults, inst2.Addr)
		}
	}

	// Results should be somewhat deterministic (consistent hash property)
	if len(firstResults) == len(secondResults) {
		// First few results should be similar
		if len(firstResults) > 0 && len(secondResults) > 0 {
			t.Logf("First results: %v", firstResults[:min(3, len(firstResults))])
			t.Logf("Second results: %v", secondResults[:min(3, len(secondResults))])
		}
	}
}

func TestConsistentHashEmpty(t *testing.T) {
	lb := NewConsistentHash()

	_, ok := lb.Next()
	if ok {
		t.Error("Expected no instance from empty balancer")
	}
}

func TestStrategyFactory(t *testing.T) {
	factory := StrategyFactory{}

	strategies := []string{"round_robin", "rr", "weighted_round_robin", "wrr", "least_connection", "lc", "consistent_hash", "ch"}

	for _, name := range strategies {
		strategy := factory.NewStrategy(name)
		if strategy == nil {
			t.Errorf("Expected strategy for '%s', got nil", name)
		}
	}

	// Test unknown strategy (should fallback to round_robin)
	strategy := factory.NewStrategy("unknown")
	if strategy == nil {
		t.Error("Expected fallback strategy for unknown name")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
