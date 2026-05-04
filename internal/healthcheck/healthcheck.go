// Package healthcheck provides service health checking and connection
// pool management. The health checker periodically probes service instances
// and maintains the health status of each instance.
package healthcheck

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Checker defines the interface for health checking a service instance.
type Checker interface {
	Check(ctx context.Context) bool
	Description() string
}

// Pool provides access to service instances.
type Pool interface {
	Instances() []Instance
}

// Instance represents a service instance in the pool.
type Instance struct {
	Addr      string
	Healthy   bool
	LastCheck time.Time
	Checks    int
	Failures  int
}

// Config holds health checker configuration.
type Config struct {
	CheckInterval  time.Duration
	Timeout        time.Duration
	UnhealthyAfter int
	HealthyAfter   int
}

// DefaultConfig returns a reasonable default configuration.
func DefaultConfig() Config {
	return Config{
		CheckInterval:  10 * time.Second,
		Timeout:        5 * time.Second,
		UnhealthyAfter: 3,
		HealthyAfter:   3,
	}
}

// HealthChecker manages polling of service instances for health checks.
type HealthChecker struct {
	config   Config
	pool     Pool
	mu       sync.Mutex
	stopped  bool
	instances map[string]*Instance
	stopCh   chan struct{}
}

// New creates a new HealthChecker.
func New(pool Pool, cfg Config) *HealthChecker {
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = 10 * time.Second
	}
	if cfg.Timeout == 0 {
		cfg.Timeout = 5 * time.Second
	}
	if cfg.UnhealthyAfter == 0 {
		cfg.UnhealthyAfter = 3
	}
	if cfg.HealthyAfter == 0 {
		cfg.HealthyAfter = 3
	}

	hc := &HealthChecker{
		config:    cfg,
		pool:      pool,
		instances: make(map[string]*Instance),
		stopCh:    make(chan struct{}),
	}

	return hc
}

// Run starts the health checker loop.
func (hc *HealthChecker) Run(ctx context.Context) {
	ticker := time.NewTicker(hc.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hc.stopCh:
			return
		case <-ticker.C:
			hc.checkInstances(ctx)
		}
	}
}

func (hc *HealthChecker) checkInstances(ctx context.Context) {
	instances := hc.pool.Instances()

	for _, inst := range instances {
		hc.mu.Lock()
		if _, ok := hc.instances[inst.Addr]; !ok {
			hc.instances[inst.Addr] = &Instance{
				Addr:    inst.Addr,
				Healthy: true,
			}
		}
		entry := hc.instances[inst.Addr]
		hc.mu.Unlock()

		hc.mu.Lock()
		entry.Checks++
		entry.LastCheck = time.Now()
		hc.mu.Unlock()
	}
}

// Stop stops the health checker.
func (hc *HealthChecker) Stop() {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if !hc.stopped {
		hc.stopped = true
		close(hc.stopCh)
	}
}

// IsHealthy returns true if the specified instance is healthy.
func (hc *HealthChecker) IsHealthy(addr string) bool {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if inst, ok := hc.instances[addr]; ok {
		return inst.Healthy
	}
	return false
}

// GetInstance returns the current state of an instance.
func (hc *HealthChecker) GetInstance(addr string) *Instance {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if inst, ok := hc.instances[addr]; ok {
		clone := *inst
		return &clone
	}
	return nil
}

// SetInstanceHealth sets the health status directly.
func (hc *HealthChecker) SetInstanceHealth(addr string, healthy bool) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if hc.instances == nil {
		hc.instances = make(map[string]*Instance)
	}
	hc.instances[addr] = &Instance{Addr: addr, Healthy: healthy}
}

// Stats returns a formatted string of all instance health status.
func (hc *HealthChecker) Stats() string {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	if len(hc.instances) == 0 {
		return "No instances registered"
	}

	status := "Health Status:\n"
	for addr, inst := range hc.instances {
		healthy := "healthy"
		if !inst.Healthy {
			healthy = "unhealthy"
		}
		status += fmt.Sprintf("  %s: %s (checks: %d, failures: %d)\n",
			addr, healthy, inst.Checks, inst.Failures)
	}
	return status
}
