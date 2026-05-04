// Package circuitbreaker implements the Google SRE circuit breaker pattern.
//
// The breaker transitions between three states:
//  - Closed: Normal operation, requests flow through
//  - Half-Open: Limited requests allowed for recovery testing
//  - Open: All requests fail fast
//
// State transitions:
//  - Closed -> Half-Open: After AllowableFailures consecutive failures
//  - Half-Open -> Closed: After MaxFailurePercent success rate met
//  - Half-Open -> Open: After any single failure
//  - Closed -> Half-Open: After Timeout of last transition
//
// Usage:
//
//	cb := New(3, 0.5, 60*time.Second, 100*time.Millisecond)
//	if !cb.Allow() {
//	    return fallback()
//	}
//	success, err := doRequest()
//	cb.Record(success, err)
package circuitbreaker

import (
	"sync"
	"time"
)

// Config holds the circuit breaker configuration.
type Config struct {
	// Maximum consecutive failures before transitioning to half-open
	AllowableFailures int

	// Success rate required to transition back to closed
	MinSuccessRate float64

	// Time after which half-open breaker attempts recovery
	Timeout time.Duration

	// Minimum time between half-open -> closed transitions
	HalfOpenTimeout time.Duration

	// Minimum window to track metrics (prevents premature closing)
	MinWindow time.Duration
}

// DefaultConfig returns a reasonable default configuration.
func DefaultConfig() Config {
	return Config{
		AllowableFailures: 3,
		MinSuccessRate:    0.5,
		Timeout:           60 * time.Second,
		HalfOpenTimeout:   100 * time.Millisecond,
		MinWindow:         10 * time.Second,
	}
}

// State represents the current breaker state.
type State int

const (
	StateClosed   State = iota // Normal operation
	StateHalfOpen              // Testing recovery
	StateOpen                  // Breaking
)

// StateError represents a circuit breaker error.
type StateError struct {
	state State
}

func (e *StateError) Error() string {
	switch e.state {
	case StateClosed:
		return "request would cause half-open -> closed transition too quickly"
	case StateHalfOpen:
		return "circuit breaker is half-open, not ready for new requests"
	default:
		return "circuit breaker is open, requests are rejected"
	}
}

// CircuitBreaker implements the Google SRE circuit breaker pattern.
type CircuitBreaker struct {
	mu                 sync.Mutex
	config             Config
	state              State
	failureCount       int
	successCount       int
	totalCount         int
	lastFailureTime    time.Time
	lastTransitionTime time.Time
	recordingWindow    bool
	startTime          time.Time
}

// New creates a new CircuitBreaker with the given configuration.
func New(config Config) *CircuitBreaker {
	if config.AllowableFailures <= 0 {
		config.AllowableFailures = 3
	}
	if config.MinSuccessRate <= 0 || config.MinSuccessRate > 1 {
		config.MinSuccessRate = 0.5
	}
	if config.Timeout <= 0 {
		config.Timeout = 60 * time.Second
	}

	return &CircuitBreaker{
		config:  config,
		state:   StateClosed,
	}
}

// Allow determines if the breaker allows the next request.
// Returns false if the breaker is in a state that should reject requests.
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Transition to half-open if timeout has elapsed
		if time.Since(cb.lastTransitionTime) >= cb.config.Timeout {
			cb.transition(StateHalfOpen, false)
			return true
		}
		return false
	case StateHalfOpen:
		// Allow limited requests in half-open state
		return cb.totalCount == 0 || cb.successCount > 0
	default:
		return false
	}
}

// Record updates breaker state based on the request outcome.
// success: true if request succeeded, false otherwise
// err: error (nil if successful)
func (cb *CircuitBreaker) Record(success bool, err error) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateClosed {
		cb.recordClosed(success, err)
	} else if cb.state == StateHalfOpen {
		cb.recordHalfOpen(success, err)
	}
}

func (cb *CircuitBreaker) recordClosed(success bool, err error) {
	if !cb.recordingWindow {
		cb.startTime = time.Now()
		cb.recordingWindow = true
	}

	cb.totalCount++

	if success {
		cb.failureCount = 0  // Reset on success

		// Check if we can transition to closed
		if cb.state == StateHalfOpen && cb.config.HalfOpenTimeout > 0 {
			if time.Since(cb.lastTransitionTime) >= cb.config.HalfOpenTimeout {
				cb.transition(StateClosed, true)
			}
		}
	} else {
		cb.failureCount++
		cb.lastFailureTime = time.Now()

		// Check if we should open (too many failures)
		if cb.failureCount >= cb.config.AllowableFailures {
			if cb.config.MinWindow > 0 && time.Since(cb.startTime) < cb.config.MinWindow {
				return // Don't transition within min window
			}
			cb.transition(StateOpen, false)
		}
	}
}

func (cb *CircuitBreaker) recordHalfOpen(success bool, err error) {
	cb.totalCount++

	if success {
		cb.successCount++
		// Check success rate
		if cb.totalCount >= 5 { // Need minimum samples
			rate := float64(cb.successCount) / float64(cb.totalCount)
			if rate >= cb.config.MinSuccessRate {
				cb.transition(StateClosed, true)
			}
		}
	} else {
		cb.failureCount++
		// Any failure in half-open -> open
		cb.transition(StateOpen, false)
	}
}

func (cb *CircuitBreaker) transition(newState State, success bool) {
	cb.state = newState
	cb.lastTransitionTime = time.Now()

	if newState == StateClosed {
		cb.failureCount = 0
		cb.successCount = 0
		cb.totalCount = 0
	} else if newState == StateHalfOpen {
		cb.successCount = 0
		cb.totalCount = 0
	}
}

// State returns the current breaker state.
func (cb *CircuitBreaker) State() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// Pending returns the number of requests that would be pending if breaker is open.
func (cb *CircuitBreaker) Pending() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state == StateOpen && time.Since(cb.lastTransitionTime) < cb.config.Timeout
}

// GoEstimate returns the time until the breaker will be open.
// Returns the estimated time based on failure count and configuration.
func (cb *CircuitBreaker) KeepTime() float64 {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		remaining := cb.config.Timeout - time.Since(cb.lastTransitionTime)
		if remaining > 0 {
			return remaining.Seconds()
		}
		return 0
	case StateHalfOpen:
		remaining := cb.config.HalfOpenTimeout - time.Since(cb.lastTransitionTime)
		if remaining > 0 {
			return remaining.Seconds()
		}
		return 0
	default:
		// Estimate based on failure progress
		if cb.config.AllowableFailures > 0 {
			progress := float64(cb.failureCount) / float64(cb.config.AllowableFailures)
			remainingTime := cb.config.Timeout * time.Duration(1-progress)
			return remainingTime.Seconds()
		}
	}
	return 0
}
