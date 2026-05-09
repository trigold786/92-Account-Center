package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

// CircuitBreaker implements a simple circuit breaker pattern.
type CircuitBreaker struct {
	mu          sync.RWMutex
	state       State
	failureCount int
	maxFailures int
	timeout     time.Duration
	lastStateChange time.Time
}

// State represents the state of the circuit breaker.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// NewCircuitBreaker creates a new circuit breaker.
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

// Execute runs the provided function with circuit breaker protection.
func (cb *CircuitBreaker) Execute(f func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastStateChange) < cb.timeout {
			return errors.New("circuit breaker is open")
		}
		cb.state = StateHalfOpen
	case StateHalfOpen:
		// Allow one request to test if the service is healthy
	}

	err := f()
	if err != nil {
		cb.onError()
		return err
	}

	cb.onSuccess()
	return nil
}

// onError increments the failure count and may open the circuit.
func (cb *CircuitBreaker) onError() {
	cb.failureCount++
	if cb.failureCount >= cb.maxFailures {
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
	}
}

// onSuccess resets the failure count and closes the circuit.
func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0
	cb.state = StateClosed
}

// State returns the current state of the circuit breaker.
func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}