package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
)

type AdEventRepository interface {
	InsertAdEvent(ctx context.Context, e *model.AdEvent) error
	GetAdMetrics(ctx context.Context, eventType, placement string, start, end time.Time) (map[string]int64, error)
}

type AdEventService struct {
	repo      AdEventRepository
	events    []*model.AdEvent
	mu        sync.RWMutex
	idCounter atomic.Int64
}

func NewAdEventService(repo AdEventRepository) *AdEventService {
	return &AdEventService{repo: repo}
}

func (s *AdEventService) TrackAdEvent(ctx context.Context, e *model.AdEvent) (*model.AdEvent, error) {
	e.ID = s.idCounter.Add(1)
	if e.Timestamp.IsZero() {
		e.Timestamp = time.Now()
	}

	if s.repo != nil {
		if err := s.repo.InsertAdEvent(ctx, e); err != nil {
			return nil, fmt.Errorf("failed to track ad event: %w", err)
		}
		return e, nil
	}

	s.mu.Lock()
	s.events = append(s.events, e)
	s.mu.Unlock()
	return e, nil
}

func (s *AdEventService) GetAdMetrics(ctx context.Context, eventType, placement string) (map[string]int64, error) {
	if s.repo != nil {
		return s.repo.GetAdMetrics(ctx, eventType, placement, time.Now().Add(-24*time.Hour), time.Now())
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := map[string]int64{
		"impressions": 0,
		"clicks":      0,
		"conversions": 0,
	}

	for _, e := range s.events {
		if eventType != "" && e.EventType != eventType {
			continue
		}
		if placement != "" && e.Placement != placement {
			continue
		}
		switch e.EventType {
		case "ad_splash_shown", "ad_banner_shown":
			metrics["impressions"]++
		case "ad_clicked":
			metrics["clicks"]++
		}
		if e.Converted {
			metrics["conversions"]++
		}
	}

	return metrics, nil
}
