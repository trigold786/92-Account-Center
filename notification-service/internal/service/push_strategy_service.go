package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
)

type PushStrategyRepository interface {
	CreateStrategy(ctx context.Context, s *model.PushStrategy) error
	GetStrategy(ctx context.Context, id int64) (*model.PushStrategy, error)
}

type PushStrategyService struct {
	repo      PushStrategyRepository
	strategies []*model.PushStrategy
	events    []*model.PushEvent
	mu        sync.RWMutex
	idCounter atomic.Int64
}

func NewPushStrategyService(repo PushStrategyRepository) *PushStrategyService {
	return &PushStrategyService{repo: repo}
}

func (s *PushStrategyService) CreateStrategy(ctx context.Context, strategy *model.PushStrategy) (*model.PushStrategy, error) {
	strategy.ID = s.idCounter.Add(1)
	strategy.CreatedAt = time.Now()
	strategy.UpdatedAt = strategy.CreatedAt

	if s.repo != nil {
		if err := s.repo.CreateStrategy(ctx, strategy); err != nil {
			return nil, fmt.Errorf("failed to create strategy: %w", err)
		}
		return strategy, nil
	}

	s.mu.Lock()
	s.strategies = append(s.strategies, strategy)
	s.mu.Unlock()
	return strategy, nil
}

func (s *PushStrategyService) EvaluateStrategy(ctx context.Context, strategyID int64, userID int64) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var strategy *model.PushStrategy
	for _, st := range s.strategies {
		if st.ID == strategyID {
			strategy = st
			break
		}
	}
	if strategy == nil {
		return false, "strategy not found"
	}

	if !strategy.Enabled {
		return false, "strategy disabled"
	}

	hour := time.Now().Hour()
	if strategy.DNDStartHour < strategy.DNDEndHour {
		if hour >= strategy.DNDStartHour && hour < strategy.DNDEndHour {
			return false, "DND hours"
		}
	} else if strategy.DNDStartHour > strategy.DNDEndHour {
		if hour >= strategy.DNDStartHour || hour < strategy.DNDEndHour {
			return false, "DND hours"
		}
	}

	var count int
	cutoff := time.Now().Add(-24 * time.Hour)
	for _, e := range s.events {
		if e.UserID == userID && e.StrategyID == strategyID && e.Timestamp.After(cutoff) {
			count++
		}
	}
	if strategy.FrequencyCap > 0 && count >= strategy.FrequencyCap {
		return false, "frequency cap exceeded"
	}

	s.events = append(s.events, &model.PushEvent{
		UserID:     userID,
		StrategyID: strategyID,
		Timestamp:  time.Now(),
	})
	if len(s.events) > 10000 {
		cutoff := time.Now().Add(-24 * time.Hour)
		filtered := s.events[:0]
		for _, e := range s.events {
			if e.Timestamp.After(cutoff) {
				filtered = append(filtered, e)
			}
		}
		s.events = filtered
	}

	return true, "allowed"
}
