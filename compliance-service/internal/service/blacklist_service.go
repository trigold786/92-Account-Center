package service

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/repository"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/svcconfig"
)

type BlacklistService interface {
	AddEntry(ctx context.Context, req *model.BlacklistEntryRequest) (*model.BlacklistEntry, error)
	CheckBlocked(ctx context.Context, entryType, entryValue string) (bool, string, error)
	RemoveEntry(ctx context.Context, entryType, entryValue string) error
	ListEntries(ctx context.Context, entryType string, limit, offset int) ([]model.BlacklistEntry, error)
}

type blacklistService struct {
	repo *repository.BlacklistRepository
	rdb  *redis.Client
	cfg  *svcconfig.ComplianceConfig
}

func NewBlacklistService(repo *repository.BlacklistRepository, rdb *redis.Client, cfg *svcconfig.ComplianceConfig) BlacklistService {
	return &blacklistService{repo: repo, rdb: rdb, cfg: cfg}
}

func (s *blacklistService) AddEntry(ctx context.Context, req *model.BlacklistEntryRequest) (*model.BlacklistEntry, error) {
	entry := &model.BlacklistEntry{
		EntryType:  req.EntryType,
		EntryValue: req.EntryValue,
		Reason:     req.Reason,
		CreatedBy:  req.CreatedBy,
	}
	if entry.CreatedBy == "" {
		entry.CreatedBy = "system"
	}
	if req.ExpiresAt != "" {
		t, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err != nil {
			return nil, fmt.Errorf("invalid expires_at format: %w", err)
		}
		entry.ExpiresAt = &t
	}
	if err := s.repo.Create(ctx, entry); err != nil {
		return nil, err
	}
	s.cacheBlocked(ctx, req.EntryType, req.EntryValue, entry)
	return entry, nil
}

func (s *blacklistService) CheckBlocked(ctx context.Context, entryType, entryValue string) (bool, string, error) {
	if s.rdb != nil {
		key := fmt.Sprintf("blacklist:%s:%s", entryType, entryValue)
		val, err := s.rdb.Get(ctx, key).Result()
		if err == nil {
			return true, val, nil
		}
	}
	entry, err := s.repo.CheckBlocked(ctx, entryType, entryValue)
	if err != nil {
		return false, "", err
	}
	if entry != nil {
		if s.rdb != nil {
			s.cacheBlocked(ctx, entryType, entryValue, entry)
		}
		return true, entry.Reason, nil
	}
	return false, "", nil
}

func (s *blacklistService) RemoveEntry(ctx context.Context, entryType, entryValue string) error {
	if err := s.repo.Remove(ctx, entryType, entryValue); err != nil {
		return err
	}
	if s.rdb != nil {
		key := fmt.Sprintf("blacklist:%s:%s", entryType, entryValue)
		s.rdb.Del(ctx, key)
	}
	return nil
}

func (s *blacklistService) ListEntries(ctx context.Context, entryType string, limit, offset int) ([]model.BlacklistEntry, error) {
	return s.repo.List(ctx, entryType, limit, offset)
}

func (s *blacklistService) cacheBlocked(ctx context.Context, entryType, entryValue string, entry *model.BlacklistEntry) {
	if s.rdb == nil {
		return
	}
	key := fmt.Sprintf("blacklist:%s:%s", entryType, entryValue)
	ttl := s.cfg.BlacklistCacheTTL
	if entry.ExpiresAt != nil {
		ttl = time.Until(*entry.ExpiresAt)
		if ttl <= 0 {
			return
		}
	}
	s.rdb.Set(ctx, key, entry.Reason, ttl)
}
