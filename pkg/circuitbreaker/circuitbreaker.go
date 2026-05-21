package circuitbreaker

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

var ErrOpen = errors.New("circuit breaker is open")

type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "closed"
	case StateOpen:
		return "open"
	case StateHalfOpen:
		return "half-open"
	default:
		return "unknown"
	}
}

type Options struct {
	MaxFailures   int
	Timeout       time.Duration
	HalfOpenMax   int
	OnStateChange func(from, to State)
}

type CircuitBreaker struct {
	mu              sync.RWMutex
	state           State
	failureCount    int
	maxFailures     int
	timeout         time.Duration
	halfOpenMax     int
	halfOpenSuccess int
	lastStateChange time.Time
	totalSuccess    atomic.Int64
	totalFailure    atomic.Int64
	onStateChange   func(from, to State)
}

func New(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return NewWithOptions(Options{
		MaxFailures: maxFailures,
		Timeout:     timeout,
		HalfOpenMax: 1,
	})
}

func NewWithOptions(opts Options) *CircuitBreaker {
	halfOpenMax := opts.HalfOpenMax
	if halfOpenMax <= 0 {
		halfOpenMax = 1
	}
	return &CircuitBreaker{
		state:         StateClosed,
		maxFailures:   opts.MaxFailures,
		timeout:       opts.Timeout,
		halfOpenMax:   halfOpenMax,
		onStateChange: opts.OnStateChange,
	}
}

func (cb *CircuitBreaker) Execute(f func() error) error {
	cb.mu.Lock()

	isHalfOpenProbe := false
	switch cb.state {
	case StateOpen:
		if time.Since(cb.lastStateChange) < cb.timeout {
			cb.mu.Unlock()
			return ErrOpen
		}
		cb.setState(StateHalfOpen)
		cb.halfOpenSuccess = 0
		fallthrough
	case StateHalfOpen:
		if cb.halfOpenMax > 0 && cb.halfOpenSuccess >= cb.halfOpenMax {
			cb.mu.Unlock()
			return ErrOpen
		}
		cb.halfOpenSuccess++
		isHalfOpenProbe = true
	}

	err := f()

	if err != nil {
		cb.failureCount++
		cb.totalFailure.Add(1)
		if cb.state == StateHalfOpen || cb.failureCount >= cb.maxFailures {
			cb.setState(StateOpen)
		}
		cb.mu.Unlock()
		return err
	}

	cb.totalSuccess.Add(1)
	if isHalfOpenProbe && cb.halfOpenSuccess >= cb.halfOpenMax {
		cb.failureCount = 0
		cb.setState(StateClosed)
	} else if !isHalfOpenProbe {
		cb.failureCount = 0
	}
	cb.mu.Unlock()
	return nil
}

func (cb *CircuitBreaker) setState(newState State) {
	old := cb.state
	cb.state = newState
	cb.lastStateChange = time.Now()
	if cb.onStateChange != nil {
		cb.onStateChange(old, newState)
	}
}

func (cb *CircuitBreaker) State() State {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) SuccessCount() int64 {
	return cb.totalSuccess.Load()
}

func (cb *CircuitBreaker) FailureCount() int64 {
	return cb.totalFailure.Load()
}

func (cb *CircuitBreaker) FailureRate() float64 {
	total := cb.totalSuccess.Load() + cb.totalFailure.Load()
	if total == 0 {
		return 0
	}
	return float64(cb.totalFailure.Load()) / float64(total)
}
