package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
)

type TemplateRepository interface {
	CreateTemplate(ctx context.Context, t *model.Template) error
	GetTemplate(ctx context.Context, id int64) (*model.Template, error)
	ListTemplates(ctx context.Context, channel string) ([]*model.Template, error)
	UpdateTemplate(ctx context.Context, t *model.Template) error
	DeleteTemplate(ctx context.Context, id int64) error
	ListSendRecords(ctx context.Context, templateID int64, offset, limit int) ([]*model.SendRecord, error)
}

type TemplateService struct {
	repo        TemplateRepository
	templates   []*model.Template
	sendRecords []*model.SendRecord
	mu          sync.RWMutex
	idCounter   atomic.Int64
}

func NewTemplateService(repo TemplateRepository) *TemplateService {
	return &TemplateService{repo: repo}
}

func (s *TemplateService) CreateTemplate(ctx context.Context, t *model.Template) (*model.Template, error) {
	t.ID = s.idCounter.Add(1)
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt

	if s.repo != nil {
		if err := s.repo.CreateTemplate(ctx, t); err != nil {
			return nil, fmt.Errorf("failed to create template: %w", err)
		}
		return t, nil
	}

	s.mu.Lock()
	s.templates = append(s.templates, t)
	s.mu.Unlock()
	return t, nil
}

func (s *TemplateService) GetTemplate(ctx context.Context, id int64) (*model.Template, error) {
	if s.repo != nil {
		return s.repo.GetTemplate(ctx, id)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, t := range s.templates {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, fmt.Errorf("template not found")
}

func (s *TemplateService) ListTemplates(ctx context.Context, channel string) ([]*model.Template, error) {
	if s.repo != nil {
		return s.repo.ListTemplates(ctx, channel)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Template
	for _, t := range s.templates {
		if channel == "" || t.Channel == channel {
			result = append(result, t)
		}
	}
	return result, nil
}

func (s *TemplateService) UpdateTemplate(ctx context.Context, t *model.Template) (*model.Template, error) {
	t.UpdatedAt = time.Now()

	if s.repo != nil {
		if err := s.repo.UpdateTemplate(ctx, t); err != nil {
			return nil, fmt.Errorf("failed to update template: %w", err)
		}
		return t, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.templates {
		if existing.ID == t.ID {
			t.CreatedAt = existing.CreatedAt
			s.templates[i] = t
			return t, nil
		}
	}
	return nil, fmt.Errorf("template not found")
}

func (s *TemplateService) DeleteTemplate(ctx context.Context, id int64) error {
	if s.repo != nil {
		return s.repo.DeleteTemplate(ctx, id)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for i, t := range s.templates {
		if t.ID == id {
			s.templates = append(s.templates[:i], s.templates[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("template not found")
}

func (s *TemplateService) ListSendRecords(ctx context.Context, templateID int64, offset, limit int) ([]*model.SendRecord, error) {
	if s.repo != nil {
		return s.repo.ListSendRecords(ctx, templateID, offset, limit)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.SendRecord
	for _, r := range s.sendRecords {
		if r.TemplateID == templateID {
			result = append(result, r)
		}
	}
	if offset >= len(result) {
		return nil, nil
	}
	end := offset + limit
	if end > len(result) {
		end = len(result)
	}
	return result[offset:end], nil
}
