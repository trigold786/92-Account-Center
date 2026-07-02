package saga

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type Status int

const (
	StatusPending     Status = 0
	StatusRunning     Status = 1
	StatusCompleted   Status = 2
	StatusFailed      Status = 3
	StatusCompensated Status = 4
)

// Saga orchestrates a series of steps with automatic compensation on failure.
type Saga struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	Steps     []*SagaStep `json:"-"`
	store     StateStore
	startedAt time.Time `json:"started_at"`
	mu        sync.Mutex
}

// New creates a Saga with the given name and optional Redis-backed state store.
func New(name string, rdb *redis.Client) *Saga {
	s := &Saga{
		Name:      name,
		Status:    StatusPending,
		Steps:     make([]*SagaStep, 0),
		startedAt: time.Now(),
	}
	if rdb != nil {
		s.store = NewRedisStore(rdb, 24*time.Hour)
	}
	return s
}

func (s *Saga) SetID(id string) {
	s.ID = id
}

func (s *Saga) AddStep(step *SagaStep) {
	s.Steps = append(s.Steps, step)
}

// Execute runs all saga steps sequentially, compensating completed steps on failure.
func (s *Saga) Execute(ctx context.Context) error {
	s.mu.Lock()
	s.Status = StatusRunning
	s.mu.Unlock()

	if err := s.persist(ctx); err != nil {
		return fmt.Errorf("persist saga: %w", err)
	}

	for i, step := range s.Steps {
		if err := step.Execute(ctx); err != nil {
			s.mu.Lock()
			s.Status = StatusFailed
			s.mu.Unlock()
			s.persist(ctx)
			s.compensate(ctx, i)
			return fmt.Errorf("step %s failed: %w", step.Name, err)
		}
		step.executed = true
		s.persist(ctx)
	}

	s.mu.Lock()
	s.Status = StatusCompleted
	s.mu.Unlock()
	s.persist(ctx)
	return nil
}

func (s *Saga) compensate(ctx context.Context, failedAtIndex int) {
	for i := failedAtIndex - 1; i >= 0; i-- {
		step := s.Steps[i]
		if step.executed && step.Compensate != nil {
			if err := step.Compensate(ctx); err != nil {
				slog.Error("saga compensation failed", "step", step.Name, "error", err)
			}
		}
	}
	s.mu.Lock()
	s.Status = StatusCompensated
	s.mu.Unlock()
	s.persist(ctx)
}

func (s *Saga) persist(ctx context.Context) error {
	if s.store == nil {
		return nil
	}
	return s.store.Save(ctx, s)
}

var (
	ErrStepExecution = errors.New("saga step execution error")
	ErrCompensation  = errors.New("saga compensation error")
	ErrStateNotFound = errors.New("saga state not found")
)
