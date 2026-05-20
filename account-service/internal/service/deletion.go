package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
)

type DeletionService interface {
	ProcessExpiredDeletions(ctx context.Context) (int, error)
}

type deletionService struct {
	userRepo        repository.UserRepository
	entitlementRepo repository.EntitlementRepository
	rdb             *redis.Client
	logger          *slog.Logger
}

func NewDeletionService(
	userRepo repository.UserRepository,
	entitlementRepo repository.EntitlementRepository,
	rdb *redis.Client,
	logger *slog.Logger,
) DeletionService {
	return &deletionService{
		userRepo:        userRepo,
		entitlementRepo: entitlementRepo,
		rdb:             rdb,
		logger:          logger,
	}
}

func (s *deletionService) ProcessExpiredDeletions(ctx context.Context) (int, error) {
	users, err := s.userRepo.GetExpiredDeletions(ctx)
	if err != nil {
		return 0, fmt.Errorf("query expired deletions: %w", err)
	}

	processed := 0
	for _, user := range users {
		if err := s.processOne(ctx, user.ID); err != nil {
			s.logger.Error("failed to process deletion",
				"user_id", user.ID,
				"error", err.Error(),
			)
			continue
		}
		processed++
		s.logger.Info("account anonymized",
			"user_id", user.ID,
		)
	}

	return processed, nil
}

func (s *deletionService) processOne(ctx context.Context, userID int64) error {
	if err := s.userRepo.AnonymizeUser(ctx, userID); err != nil {
		return fmt.Errorf("anonymize user %d: %w", userID, err)
	}

	if err := s.entitlementRepo.DeleteByUserID(ctx, userID); err != nil {
		s.logger.Warn("failed to delete entitlements",
			"user_id", userID,
			"error", err.Error(),
		)
	}

	if s.rdb != nil {
		sessionPattern := fmt.Sprintf("session:*:%d", userID)
		s.cleanupRedis(ctx, sessionPattern)
		cachePattern := fmt.Sprintf("entitlements:%d:*", userID)
		s.cleanupRedis(ctx, cachePattern)
	}

	s.logger.Info("deletion audit",
		"user_id", userID,
		"timestamp", time.Now().Format(time.RFC3339),
		"action", "account_anonymized",
	)

	return nil
}

func (s *deletionService) cleanupRedis(ctx context.Context, pattern string) {
	iter := s.rdb.Scan(ctx, 0, pattern, 100).Iterator()
	for iter.Next(ctx) {
		if err := s.rdb.Del(ctx, iter.Val()).Err(); err != nil {
			s.logger.Warn("redis key deletion failed",
				"key", iter.Val(),
				"error", err.Error(),
			)
		}
	}
	if err := iter.Err(); err != nil {
		s.logger.Warn("redis scan error",
			"pattern", pattern,
			"error", err.Error(),
		)
	}
}
