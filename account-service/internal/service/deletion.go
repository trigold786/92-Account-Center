package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
	circuitbreaker "github.com/trigold786/92-Account-Center/pkg/circuitbreaker"
)

type DeletionService interface {
	ProcessExpiredDeletions(ctx context.Context) (int, error)
}

type SessionInvalidator interface {
	InvalidateAllSessions(ctx context.Context, userID int64) error
}

type deletionService struct {
	userRepo           repository.UserRepository
	entitlementRepo    repository.EntitlementRepository
	rdb                *redis.Client
	logger             *slog.Logger
	sessionInvalidator SessionInvalidator
}

func NewDeletionService(
	userRepo repository.UserRepository,
	entitlementRepo repository.EntitlementRepository,
	rdb *redis.Client,
	logger *slog.Logger,
	opts ...SessionInvalidator,
) DeletionService {
	svc := &deletionService{
		userRepo:        userRepo,
		entitlementRepo: entitlementRepo,
		rdb:             rdb,
		logger:          logger,
	}
	if len(opts) > 0 {
		svc.sessionInvalidator = opts[0]
	}
	return svc
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

	if err := s.userRepo.AnonymizeEnterprisePII(ctx, userID); err != nil {
		s.logger.Warn("failed to anonymize enterprise PII",
			"user_id", userID,
			"error", err.Error(),
		)
	}

	if err := s.entitlementRepo.DeleteByUserID(ctx, userID); err != nil {
		s.logger.Warn("failed to delete entitlements",
			"user_id", userID,
			"error", err.Error(),
		)
	}

	if s.rdb != nil {
		// Delete user_sessions set (indexes live session keys)
		s.cleanupRedis(ctx, fmt.Sprintf("user_sessions:%d", userID))
		// Delete entitlement cache (correct singular key)
		s.cleanupRedis(ctx, fmt.Sprintf("entitlement:%d", userID))
	}

	if s.sessionInvalidator != nil {
		if err := s.sessionInvalidator.InvalidateAllSessions(ctx, userID); err != nil {
			s.logger.Warn("failed to invalidate sessions",
				"user_id", userID,
				"error", err.Error(),
			)
		}
	}

	auditDetails := map[string]interface{}{
		"action":                    "account_anonymized",
		"anonymized_fields":         []string{"phone_number", "account_id", "email", "password_hash", "mfa"},
		"enterprise_pii_anonymized": true,
	}
	if err := s.userRepo.WriteDeletionAudit(ctx, userID, auditDetails); err != nil {
		s.logger.Error("failed to write deletion audit log",
			"user_id", userID,
			"error", err.Error(),
		)
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

type HTTPSessionInvalidator struct {
	baseURL string
	client  *http.Client
}

func NewHTTPSessionInvalidator(baseURL string) *HTTPSessionInvalidator {
	return &HTTPSessionInvalidator{baseURL: baseURL, client: circuitbreaker.WrapHTTPClient(&http.Client{Timeout: 5 * time.Second}, "auth-service")}
}

func (h *HTTPSessionInvalidator) InvalidateAllSessions(ctx context.Context, userID int64) error {
	body, _ := json.Marshal(map[string]int64{"user_id": userID})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, h.baseURL+"/api/v1/sessions/invalidate-all", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("auth-service session invalidation returned %d", resp.StatusCode)
	}
	return nil
}
