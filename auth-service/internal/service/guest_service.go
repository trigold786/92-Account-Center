package service

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type GuestRepository interface {
	CreateGuest(ctx context.Context, guest *model.GuestSession) error
	GetByAccountID(ctx context.Context, accountID string) (*model.GuestSession, error)
	UpdateGuest(ctx context.Context, guest *model.GuestSession) error
}

type GuestService struct {
	repo      GuestRepository
	idCounter atomic.Int64
}

func NewGuestService(repo GuestRepository) *GuestService {
	return &GuestService{repo: repo}
}

func (s *GuestService) CreateGuest(ctx context.Context, deviceID string) (*model.GuestSession, error) {
	id := s.idCounter.Add(1)
	accountID := fmt.Sprintf("guest_%d_%d", id, time.Now().UnixNano())

	guest := &model.GuestSession{
		AccountID: accountID,
		DeviceID:  deviceID,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if s.repo != nil {
		if err := s.repo.CreateGuest(ctx, guest); err != nil {
			return nil, fmt.Errorf("failed to create guest: %w", err)
		}
	}

	return guest, nil
}

func (s *GuestService) UpgradeGuest(ctx context.Context, req *model.UpgradeGuestRequest) (*model.GuestSession, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("repository not available")
	}

	guest, err := s.repo.GetByAccountID(ctx, req.AccountID)
	if err != nil {
		return nil, fmt.Errorf("guest not found: %w", err)
	}

	if req.Email != "" {
		guest.Email = req.Email
	}
	if req.Phone != "" {
		guest.Phone = req.Phone
	}
	guest.Status = "upgraded"
	guest.UpdatedAt = time.Now()

	if err := s.repo.UpdateGuest(ctx, guest); err != nil {
		return nil, fmt.Errorf("failed to upgrade guest: %w", err)
	}

	return guest, nil
}
