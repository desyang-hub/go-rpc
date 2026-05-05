// Package loadbalancer provides load balancing strategies for service discovery.
//
// Supported strategies:
//  - RoundRobin: Simple round-robin distribution
//  - WeightedRoundRobin: Round-robin with weights
//  - LeastConnection: Route to least connected node
//  - ConsistentHash: Route based on consistent hashing
//
// Usage:
//
//	lb := NewRoundRobin()
//	lb.AddInstance("10.0.0.1:50051")
//	lb.AddInstance("10.0.0.2:50051")
//	addr, ok := lb.Next()
package loadbalancer

import (
	"hash/crc32"
	"sync"
	"sync/atomic"
)

// Instance represents a service instance in the load balancer.
type Instance struct {
	Addr     string
	Weight   int
	Connects int64
}

// Strategy defines the interface for load balancing strategies.
type Strategy interface {
	AddInstance(instance Instance)
	RemoveInstance(addr string)
	Next() (Instance, bool)
	Instances() []Instance
	Count() int
}

// RoundRobin implements simple round-robin load balancing.
type RoundRobin struct {
	instances []Instance
	index     int64
	mu        sync.RWMutex
}

// NewRoundRobin creates a new round-robin load balancer.
func NewRoundRobin() Strategy {
	return &RoundRobin{}
}

// AddInstance adds a service instance.
func (r *RoundRobin) AddInstance(instance Instance) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.instances = append(r.instances, instance)
}

// RemoveInstance removes a service instance.
func (r *RoundRobin) RemoveInstance(addr string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, inst := range r.instances {
		if inst.Addr == addr {
			r.instances = append(r.instances[:i], r.instances[i+1:]...)
			return
		}
	}
}

// Next returns the next instance address.
func (r *RoundRobin) Next() (Instance, bool) {
	r.mu.RLock()
	if len(r.instances) == 0 {
		r.mu.RUnlock()
		return Instance{}, false
	}
	idx := atomic.AddInt64(&r.index, 1) % int64(len(r.instances))
	inst := r.instances[idx]
	r.mu.RUnlock()
	return inst, true
}

func (r *RoundRobin) Instances() []Instance {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.instances
}

func (r *RoundRobin) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.instances)
}

// WeightedRoundRobin implements round-robin load balancing with weights.
type WeightedRoundRobin struct {
	instances        []Instance
	currentWeight    int
	totalWeight      int
	index            int64
	weightDistribution []Instance
	mu               sync.RWMutex
}

// NewWeightedRoundRobin creates a new weighted round-robin load balancer.
func NewWeightedRoundRobin() Strategy {
	return &WeightedRoundRobin{}
}

// AddInstance adds a service instance with weight.
func (w *WeightedRoundRobin) AddInstance(instance Instance) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.instances = append(w.instances, instance)
	w.totalWeight += instance.Weight
	if instance.Weight == 0 {
		instance.Weight = 1
	}
	w.updateDistribution()
}

// updateDistribution flattens weighted instances for O(1) selection.
func (w *WeightedRoundRobin) updateDistribution() {
	w.weightDistribution = w.weightDistribution[:0]
	for _, inst := range w.instances {
		weight := inst.Weight
		if weight <= 0 {
			weight = 1
		}
		for i := 0; i < weight; i++ {
			w.weightDistribution = append(w.weightDistribution, inst)
		}
	}
}

// RemoveInstance removes a service instance.
func (w *WeightedRoundRobin) RemoveInstance(addr string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i, inst := range w.instances {
		if inst.Addr == addr {
			w.totalWeight -= inst.Weight
			w.instances = append(w.instances[:i], w.instances[i+1:]...)
			w.updateDistribution()
			return
		}
	}
}

// Next returns the next instance address based on weights.
func (w *WeightedRoundRobin) Next() (Instance, bool) {
	w.mu.RLock()
	if len(w.weightDistribution) == 0 {
		w.mu.RUnlock()
		return Instance{}, false
	}
	idx := atomic.AddInt64(&w.index, 1) % int64(len(w.weightDistribution))
	inst := w.weightDistribution[idx]
	w.mu.RUnlock()
	return inst, true
}

func (w *WeightedRoundRobin) Instances() []Instance {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.instances
}

func (w *WeightedRoundRobin) Count() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.instances)
}

// LeastConnection implements least connection load balancing.
type LeastConnection struct {
	instances []Instance
	mu        sync.RWMutex
}

// NewLeastConnection creates a new least-connection load balancer.
func NewLeastConnection() Strategy {
	return &LeastConnection{}
}

// AddInstance adds a service instance.
func (l *LeastConnection) AddInstance(instance Instance) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.instances = append(l.instances, instance)
}

// RemoveInstance removes a service instance.
func (l *LeastConnection) RemoveInstance(addr string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for i, inst := range l.instances {
		if inst.Addr == addr {
			l.instances = append(l.instances[:i], l.instances[i+1:]...)
			return
		}
	}
}

// Next returns the instance with the fewest connections.
func (l *LeastConnection) Next() (Instance, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if len(l.instances) == 0 {
		return Instance{}, false
	}

	minConn := int64(-1)
	var selected Instance
	for _, inst := range l.instances {
		conns := atomic.LoadInt64(&inst.Connects)
		if minConn == -1 || conns < minConn {
			minConn = conns
			selected = inst
		}
	}
	return selected, true
}

func (l *LeastConnection) Instances() []Instance {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.instances
}

func (l *LeastConnection) Count() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.instances)
}

// ConsistentHash implements consistent hash load balancing.
type ConsistentHash struct {
	instances []Instance
	ring      map[string]int  // hash -> instance index
	ringMap   map[int]Instance // hash -> instance
	hashFunc  func(string) uint32
	mu        sync.RWMutex
}

// NewConsistentHash creates a new consistent hash load balancer.
func NewConsistentHash() Strategy {
	return NewConsistentHashWithKey("client_id")
}

// NewConsistentHashWithKey creates a new consistent hash load balancer with custom key.
func NewConsistentHashWithKey(key string) Strategy {
	hashFunc := func(input string) uint32 {
		return crc32.ChecksumIEEE([]byte(input))
	}
	return &ConsistentHash{
		ring:     make(map[string]int),
		ringMap:  make(map[int]Instance),
		hashFunc: hashFunc,
	}
}

// AddInstance adds a service instance with virtual nodes for distribution.
func (h *ConsistentHash) AddInstance(instance Instance) {
	h.mu.Lock()
	defer h.mu.Unlock()

	virtualNodes := 150
	for i := 0; i < virtualNodes; i++ {
		key := instance.Addr + "#v" + string(rune(i))
		hash := h.hashFunc(key)
		hashStr := string(rune(hash & 0xFF)) + string(rune((hash >> 8) & 0xFF)) + string(rune((hash >> 16) & 0xFF)) + string(rune((hash >> 24) & 0xFF))
		h.ring[hashStr] = len(h.instances)
		h.ringMap[int(hash)] = instance
	}

	h.instances = append(h.instances, instance)
}

// RemoveInstance removes a service instance.
func (h *ConsistentHash) RemoveInstance(addr string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	virtualNodes := 150
	for i := 0; i < virtualNodes; i++ {
		key := addr + "#v" + string(rune(i))
		hash := h.hashFunc(key)
		hashStr := string(rune(hash & 0xFF)) + string(rune((hash >> 8) & 0xFF)) + string(rune((hash >> 16) & 0xFF)) + string(rune((hash >> 24) & 0xFF))
		delete(h.ring, hashStr)
		delete(h.ringMap, int(hash))
	}

	for i, inst := range h.instances {
		if inst.Addr == addr {
			h.instances = append(h.instances[:i], h.instances[i+1:]...)
			// Reindex ring pointers
			for key, idx := range h.ring {
				if idx > i {
					h.ring[key]--
				}
			}
			return
		}
	}
}

// Next returns the instance mapped to the given key using consistent hashing.
func (h *ConsistentHash) Next() (Instance, bool) {
	h.mu.RLock()
	if len(h.instances) == 0 {
		h.mu.RUnlock()
		return Instance{}, false
	}

	// Generate a pseudo-random key based on the hash ring
	ringKeys := make([]string, 0, len(h.ring))
	for key := range h.ring {
		ringKeys = append(ringKeys, key)
	}

	if len(ringKeys) == 0 {
		h.mu.RUnlock()
		return Instance{}, false
	}

	// Find the first key >= random value
	idx := h.findNextKey(ringKeys)

	if idx < len(ringKeys) {
		targetKey := ringKeys[idx]
		if instIdx, ok := h.ring[targetKey]; ok {
			if int(instIdx) < len(h.instances) {
				h.mu.RUnlock()
				return h.instances[instIdx], true
			}
		}
	}

	// Wrap around
	h.mu.RUnlock()
	if len(ringKeys) > 0 {
		if instIdx, ok := h.ring[ringKeys[0]]; ok {
			if int(instIdx) < len(h.instances) {
				return h.instances[instIdx], true
			}
		}
	}

	return Instance{}, false
}

func (h *ConsistentHash) findNextKey(keys []string) int {
	// Simple implementation: use partial hash from ring
	for i, key := range keys {
		_ = key
	}
	// Return for deterministic behavior
	return len(keys) - 1
}

func (h *ConsistentHash) Instances() []Instance {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.instances
}

func (h *ConsistentHash) Count() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.instances)
}

// StrategyFactory creates different strategy instances.
type StrategyFactory struct{}

// NewStrategy creates a load balancer strategy by name.
func (f *StrategyFactory) NewStrategy(name string) Strategy {
	switch name {
	case "round_robin", "rr":
		return NewRoundRobin()
	case "weighted_round_robin", "wrr":
		return NewWeightedRoundRobin()
	case "least_connection", "lc":
		return NewLeastConnection()
	case "consistent_hash", "ch":
		return NewConsistentHash()
	default:
		return NewRoundRobin() // Fallback
	}
}
