package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/trigold786/92-Account-Center/account-service/internal/cache"
	"github.com/trigold786/92-Account-Center/account-service/internal/model"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
)

var (
	ErrEntitlementNotFound = errors.New("entitlement not found")
	ErrInsufficientQuota   = errors.New("insufficient quota")
)

var tierEntitlements = map[int]map[string]int{
	2: {"api_call_limit": 1000, "storage_space": 1024},
	3: {"api_call_limit": 5000, "storage_space": 5120},
	4: {"api_call_limit": 50000, "storage_space": 51200},
}

type EntitlementService interface {
	GetUserEntitlements(ctx context.Context, userID int64) ([]model.Entitlement, error)
	ConsumeQuota(ctx context.Context, userID int64, featureCode string, amount int) (*model.ConsumeResponse, error)
	GrantEntitlements(ctx context.Context, userID int64, tierLevel int) error
	DeleteUserEntitlements(ctx context.Context, userID int64) error
}

type entitlementService struct {
	repo  repository.EntitlementRepository
	cache *cache.EntitlementCache
}

func NewEntitlementService(repo repository.EntitlementRepository, cache *cache.EntitlementCache) EntitlementService {
	return &entitlementService{repo: repo, cache: cache}
}

func (s *entitlementService) GetUserEntitlements(ctx context.Context, userID int64) ([]model.Entitlement, error) {
	entitlements, err := s.repo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return entitlements, nil
}

func (s *entitlementService) ConsumeQuota(ctx context.Context, userID int64, featureCode string, amount int) (*model.ConsumeResponse, error) {
	ok, err := s.cache.ConsumeQuota(ctx, userID, featureCode, amount)
	if err != nil {
		e, dbErr := s.repo.GetByUserAndFeature(ctx, userID, featureCode)
		if dbErr != nil {
			return nil, dbErr
		}
		if e == nil {
			return nil, ErrEntitlementNotFound
		}
		if e.TotalQuota-e.UsedQuota < amount {
			return &model.ConsumeResponse{Success: false, Remaining: e.TotalQuota - e.UsedQuota}, ErrInsufficientQuota
		}
		newUsed := e.UsedQuota + amount
		if err := s.repo.UpdateQuota(ctx, e.ID, newUsed); err != nil {
			return nil, err
		}
		entitlements, _ := s.repo.GetByUserID(ctx, userID)
		if len(entitlements) > 0 {
			if err := s.cache.WarmCache(ctx, entitlements); err != nil {
				slog.Warn("failed to warm entitlement cache", "user_id", userID, "error", err)
			}
		}
		return &model.ConsumeResponse{Success: true, Remaining: e.TotalQuota - newUsed}, nil
	}
	if !ok {
		return &model.ConsumeResponse{Success: false}, ErrInsufficientQuota
	}

	e, err := s.repo.GetByUserAndFeature(ctx, userID, featureCode)
	if err == nil && e != nil {
		if err := s.repo.UpdateQuota(ctx, e.ID, e.UsedQuota+amount); err != nil {
			slog.Warn("failed to update quota in repo after cache consume", "user_id", userID, "feature", featureCode, "error", err)
		}
	}

	quota, _ := s.cache.GetQuota(ctx, userID, featureCode)
	remaining := 0
	if quota != nil {
		remaining = quota.Total - quota.Used
	}
	return &model.ConsumeResponse{Success: true, Remaining: remaining}, nil
}

func (s *entitlementService) DeleteUserEntitlements(ctx context.Context, userID int64) error {
	if err := s.repo.DeleteByUserID(ctx, userID); err != nil {
		return err
	}
	if s.cache != nil {
		if err := s.cache.InvalidateCache(ctx, userID); err != nil {
			slog.Warn("failed to invalidate entitlement cache", "user_id", userID, "error", err)
		}
	}
	return nil
}

func (s *entitlementService) GrantEntitlements(ctx context.Context, userID int64, tierLevel int) error {
	features, ok := tierEntitlements[tierLevel]
	if !ok {
		return fmt.Errorf("no entitlements defined for tier %d", tierLevel)
	}

	existing, _ := s.repo.GetByUserID(ctx, userID)
	existingMap := make(map[string]*model.Entitlement)
	for i := range existing {
		existingMap[existing[i].FeatureCode] = &existing[i]
	}

	for featureCode, total := range features {
		if e, exists := existingMap[featureCode]; exists {
			if err := s.repo.UpdateTotalQuota(ctx, e.ID, total); err != nil {
				return err
			}
		} else {
			e := &model.Entitlement{
				UserID:      userID,
				FeatureCode: featureCode,
				TotalQuota:  total,
				UsedQuota:   0,
			}
			if err := s.repo.Create(ctx, e); err != nil {
				return err
			}
		}
		if err := s.cache.GrantQuota(ctx, userID, featureCode, total); err != nil {
			slog.Warn("failed to grant quota in cache", "user_id", userID, "feature", featureCode, "error", err)
		}
	}

	return nil
}
