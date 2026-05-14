package circuitbreaker

import (
	"errors"
	"sync"
	"time"
)

type CircuitBreaker struct {
	mu               sync.RWMutex
	state            State
	failureCount     int
	maxFailures      int
	timeout          time.Duration
	lastStateChange  time.Time
}

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func New(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return NewCircuitBreaker(maxFailures, timeout)
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       StateClosed,
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

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
	}

	err := f()
	if err != nil {
		cb.onError()
		return err
	}

	cb.onSuccess()
	return nil
}

func (cb *CircuitBreaker) onError() {
	cb.failureCount++
	if cb.failureCount >= cb.maxFailures {
		cb.state = StateOpen
		cb.lastStateChange = time.Now()
	}
}

func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0
	cb.state = StateClosed
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}
