package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
)

type ABTestStore interface {
	SaveExperiment(ctx context.Context, exp *model.ABTest) error
	GetExperiment(ctx context.Context, id string) (*model.ABTest, error)
	SaveAssignment(ctx context.Context, assignment *model.ABAssignment) error
	GetAssignment(ctx context.Context, experimentID, userID string) (*model.ABAssignment, error)
	SaveEvent(ctx context.Context, event *model.ABEvent) error
	GetEvents(ctx context.Context, experimentID string) ([]model.ABEvent, error)
}

type ABTestService struct {
	mu           sync.RWMutex
	experiments  map[string]*model.ABTest
	assignments  map[string]*model.ABAssignment
	events       []model.ABEvent
	store        ABTestStore
}

func NewABTestService() *ABTestService {
	return &ABTestService{
		experiments: make(map[string]*model.ABTest),
		assignments: make(map[string]*model.ABAssignment),
	}
}

func (s *ABTestService) CreateExperiment(ctx context.Context, name string, variants []model.ABVariant) (*model.ABTest, error) {
	if name == "" {
		return nil, fmt.Errorf("experiment name is required")
	}
	if len(variants) < 2 {
		return nil, fmt.Errorf("at least 2 variants required")
	}

	exp := &model.ABTest{
		ID:        uuid.New().String(),
		Name:      name,
		Variants:  variants,
		Status:    "active",
		CreatedAt: time.Now(),
	}

	s.mu.Lock()
	s.experiments[exp.ID] = exp
	s.mu.Unlock()

	return exp, nil
}

func (s *ABTestService) AssignVariant(ctx context.Context, experimentID, userID string) (*model.ABAssignment, error) {
	key := experimentID + ":" + userID

	s.mu.RLock()
	if existing, ok := s.assignments[key]; ok {
		s.mu.RUnlock()
		return existing, nil
	}
	s.mu.RUnlock()

	s.mu.RLock()
	exp, ok := s.experiments[experimentID]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}

	variant := s.deterministicAssign(exp, userID)
	assignment := &model.ABAssignment{
		ExperimentID: experimentID,
		UserID:       userID,
		Variant:      variant,
	}

	s.mu.Lock()
	s.assignments[key] = assignment
	s.mu.Unlock()

	return assignment, nil
}

func (s *ABTestService) deterministicAssign(exp *model.ABTest, userID string) string {
	hash := sha256.Sum256([]byte(exp.ID + ":" + userID))
	hashStr := hex.EncodeToString(hash[:])
	var hashVal uint64
	for i := 0; i < 16 && i < len(hashStr); i++ {
		hashVal = hashVal*16 + uint64(hexDigit(hashStr[i]))
	}

	totalWeight := 0.0
	for _, v := range exp.Variants {
		totalWeight += v.Weight
	}

	threshold := float64(hashVal%10000) / 10000.0 * totalWeight
	cumulative := 0.0
	for _, v := range exp.Variants {
		cumulative += v.Weight
		if threshold < cumulative {
			return v.Name
		}
	}
	return exp.Variants[len(exp.Variants)-1].Name
}

func hexDigit(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'a' && c <= 'f' {
		return int(c-'a') + 10
	}
	return 0
}

func (s *ABTestService) RecordEvent(ctx context.Context, experimentID, userID, variant, eventType string) error {
	event := model.ABEvent{
		ExperimentID: experimentID,
		UserID:       userID,
		Variant:      variant,
		EventType:    eventType,
		Timestamp:    time.Now().Format(time.RFC3339),
	}
	s.mu.Lock()
	s.events = append(s.events, event)
	s.mu.Unlock()
	return nil
}

func (s *ABTestService) GetResults(ctx context.Context, experimentID string) (*model.ABTestResult, error) {
	s.mu.RLock()
	exp, ok := s.experiments[experimentID]
	s.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("experiment not found: %s", experimentID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	variantStats := make(map[string]*model.ABVariantResult)
	for _, v := range exp.Variants {
		variantStats[v.Name] = &model.ABVariantResult{Variant: v.Name}
	}

	for _, evt := range s.events {
		if evt.ExperimentID != experimentID {
			continue
		}
		stats, ok := variantStats[evt.Variant]
		if !ok {
			continue
		}
		if evt.EventType == "impression" || evt.EventType == "conversion" {
			stats.Count++
			if evt.EventType == "conversion" {
				stats.Conversions++
			}
		}
	}

	results := make([]model.ABVariantResult, 0, len(variantStats))
	for _, v := range exp.Variants {
		stats := variantStats[v.Name]
		if stats.Count > 0 {
			stats.ConversionRate = float64(stats.Conversions) / float64(stats.Count)
			stats.Confidence = s.calculateConfidence(stats.Count, stats.Conversions)
		}
		results = append(results, *stats)
	}

	return &model.ABTestResult{
		ExperimentID: experimentID,
		Variants:     results,
	}, nil
}

func (s *ABTestService) calculateConfidence(count, conversions int) float64 {
	if count == 0 {
		return 0
	}
	p := float64(conversions) / float64(count)
	n := float64(count)
	se := math.Sqrt(p * (1 - p) / n)
	z := 1.96
	return math.Min(1.0, se*z*10)
}
