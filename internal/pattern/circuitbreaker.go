package pattern

import (
	"sync"
	"sync/atomic"
	"time"
)

// State represents the circuit breaker state.
type State int32

const (
	Closed State = iota
	Open
	HalfOpen
)

// CircuitBreaker manages the state of a single backend.
type CircuitBreaker struct {
	state            int32
	failureCount     int32
	failureThreshold int32
	cooldown         time.Duration
	lastFailure      time.Time
	mu               sync.Mutex
}

// New creates a new CircuitBreaker.
func New(failureThreshold int32, cooldown time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            int32(Closed),
		failureThreshold: failureThreshold,
		cooldown:         cooldown,
	}
}

// IsAvailable checks if the backend is available for requests.
func (cb *CircuitBreaker) IsAvailable() bool {
	state := State(atomic.LoadInt32(&cb.state))
	if state == Closed {
		return true
	}
	if state == Open {
		cb.mu.Lock()
		defer cb.mu.Unlock()
		if time.Since(cb.lastFailure) > cb.cooldown {
			atomic.StoreInt32(&cb.state, int32(HalfOpen))
			return true
		}
		return false
	}
	// HalfOpen: Allow one request to test
	return true
}

// RecordSuccess marks a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	if State(atomic.LoadInt32(&cb.state)) == HalfOpen {
		cb.mu.Lock()
		defer cb.mu.Unlock()
		atomic.StoreInt32(&cb.state, int32(Closed))
		atomic.StoreInt32(&cb.failureCount, 0)
	}
}

// RecordFailure marks a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	atomic.AddInt32(&cb.failureCount, 1)
	if atomic.LoadInt32(&cb.failureCount) >= cb.failureThreshold {
		cb.mu.Lock()
		defer cb.mu.Unlock()
		atomic.StoreInt32(&cb.state, int32(Open))
		cb.lastFailure = time.Now()
	}
}
