package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

type FAQRepository interface {
	ListFAQs(ctx context.Context, category string) ([]*model.FAQ, error)
	SearchFAQs(ctx context.Context, query string) ([]*model.FAQ, error)
	CreateFAQ(ctx context.Context, faq *model.FAQ) error
	UpdateFAQ(ctx context.Context, faq *model.FAQ) error
}

type FAQService struct {
	repo      FAQRepository
	faqs      []*model.FAQ
	mu        sync.RWMutex
	idCounter atomic.Int64
}

func NewFAQService(repo FAQRepository) *FAQService {
	return &FAQService{repo: repo}
}

func (s *FAQService) ListFAQs(ctx context.Context, category string) ([]*model.FAQ, error) {
	if s.repo != nil {
		return s.repo.ListFAQs(ctx, category)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.FAQ
	for _, f := range s.faqs {
		if category == "" || f.Category == category {
			result = append(result, f)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].SortOrder < result[j].SortOrder
	})

	return result, nil
}

func (s *FAQService) SearchFAQs(ctx context.Context, query string) ([]*model.FAQ, error) {
	if s.repo != nil {
		return s.repo.SearchFAQs(ctx, query)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	queryLower := strings.ToLower(query)
	var result []*model.FAQ
	for _, f := range s.faqs {
		if strings.Contains(strings.ToLower(f.Question), queryLower) ||
			strings.Contains(strings.ToLower(f.Answer), queryLower) ||
			strings.Contains(strings.ToLower(f.Tags), queryLower) {
			result = append(result, f)
		}
	}

	return result, nil
}

func (s *FAQService) CreateFAQ(ctx context.Context, faq *model.FAQ) (*model.FAQ, error) {
	faq.ID = s.idCounter.Add(1)

	if s.repo != nil {
		if err := s.repo.CreateFAQ(ctx, faq); err != nil {
			return nil, fmt.Errorf("failed to create FAQ: %w", err)
		}
		return faq, nil
	}

	s.mu.Lock()
	s.faqs = append(s.faqs, faq)
	s.mu.Unlock()
	return faq, nil
}

func (s *FAQService) UpdateFAQ(ctx context.Context, faq *model.FAQ) (*model.FAQ, error) {
	if s.repo != nil {
		if err := s.repo.UpdateFAQ(ctx, faq); err != nil {
			return nil, fmt.Errorf("failed to update FAQ: %w", err)
		}
		return faq, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, f := range s.faqs {
		if f.ID == faq.ID {
			s.faqs[i] = faq
			return faq, nil
		}
	}

	return nil, fmt.Errorf("FAQ not found")
}
