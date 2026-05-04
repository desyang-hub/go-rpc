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
}

// NewRoundRobin creates a new round-robin load balancer.
func NewRoundRobin() Strategy {
	return &RoundRobin{}
}

// AddInstance adds a service instance.
func (r *RoundRobin) AddInstance(instance Instance) {
	r.instances = append(r.instances, instance)
}

// RemoveInstance removes a service instance.
func (r *RoundRobin) RemoveInstance(addr string) {
	for i, inst := range r.instances {
		if inst.Addr == addr {
			r.instances = append(r.instances[:i], r.instances[i+1:]...)
			return
		}
	}
}

// Next returns the next instance address.
func (r *RoundRobin) Next() (Instance, bool) {
	if len(r.instances) == 0 {
		return Instance{}, false
	}
	idx := atomic.AddInt64(&r.index, 1) % int64(len(r.instances))
	return r.instances[idx], true
}

func (r *RoundRobin) Instances() []Instance {
	return r.instances
}

func (r *RoundRobin) Count() int {
	return len(r.instances)
}
