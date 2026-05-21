package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
)

type StreamRepository interface {
	InsertEvent(ctx context.Context, e *model.StreamEvent) error
	GetEventsByTimeRange(ctx context.Context, start, end time.Time) ([]*model.StreamEvent, error)
}

type StreamService struct {
	repo      StreamRepository
	events    []*model.StreamEvent
	mu        sync.RWMutex
	idCounter atomic.Int64
}

func NewStreamService(repo StreamRepository) *StreamService {
	return &StreamService{repo: repo}
}

func (s *StreamService) ProcessEvent(ctx context.Context, e *model.StreamEvent) (*model.StreamEvent, error) {
	e.ID = s.idCounter.Add(1)
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}

	if s.repo != nil {
		if err := s.repo.InsertEvent(ctx, e); err != nil {
			return nil, fmt.Errorf("failed to process event: %w", err)
		}
		return e, nil
	}

	s.mu.Lock()
	s.events = append(s.events, e)
	if len(s.events) > 50000 {
		cutoff := time.Now().Add(-1 * time.Hour)
		filtered := s.events[:0]
		for _, ev := range s.events {
			if ev.Timestamp.After(cutoff) {
				filtered = append(filtered, ev)
			}
		}
		s.events = filtered
	}
	s.mu.Unlock()
	return e, nil
}

func (s *StreamService) GetOnlineCount(ctx context.Context, window time.Duration) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := time.Now().Add(-window)
	seen := make(map[int64]bool)
	for _, e := range s.events {
		if e.Timestamp.After(cutoff) {
			seen[e.UserID] = true
		}
	}
	return int64(len(seen)), nil
}

type FunnelStep struct {
	Step string `json:"step"`
	Count int64 `json:"count"`
}

func (s *StreamService) GetRealtimeFunnel(ctx context.Context, window time.Duration, steps []string) ([]FunnelStep, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := time.Now().Add(-window)
	stepUsers := make([]map[int64]bool, len(steps))
	for i := range stepUsers {
		stepUsers[i] = make(map[int64]bool)
	}

	for _, e := range s.events {
		if !e.Timestamp.After(cutoff) {
			continue
		}
		for i, step := range steps {
			if e.EventType == step {
				stepUsers[i][e.UserID] = true
			}
		}
	}

	result := make([]FunnelStep, len(steps))
	for i, step := range steps {
		result[i] = FunnelStep{
			Step:  step,
			Count: int64(len(stepUsers[i])),
		}
	}
	return result, nil
}
