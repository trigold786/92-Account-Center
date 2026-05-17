package service

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/repository"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/compliance-service/pkg/crypto"
)

var (
	ErrLogNotFound      = errors.New("audit log not found")
	ErrInvalidRetention = errors.New("retention days must be positive")
)

type AuditService interface {
	RecordLog(ctx context.Context, entry *model.AuditLogEntry) (*model.AuditLog, error)
	RecordBatch(ctx context.Context, entries []model.AuditLogEntry) (*model.BatchAuditLogResponse, error)
	GetLogsByUser(ctx context.Context, userID string, limit, offset int) ([]*model.AuditLog, error)
	GetLogsByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*model.AuditLog, error)
	VerifyLogIntegrity(ctx context.Context, logID string) (*model.IntegrityVerifyResponse, error)
	CleanupOldLogs(ctx context.Context, retentionDays int) (*model.CleanupResponse, error)
}

type auditService struct {
	repo repository.AuditRepository
	cfg  *svcconfig.ComplianceConfig
}

func NewAuditService(repo repository.AuditRepository, cfg *svcconfig.ComplianceConfig) AuditService {
	return &auditService{repo: repo, cfg: cfg}
}

func (s *auditService) RecordLog(ctx context.Context, entry *model.AuditLogEntry) (*model.AuditLog, error) {
	eventTime := time.Now()
	if entry.EventTime != nil {
		eventTime = *entry.EventTime
	}

	var userID *int64
	if entry.UserID != nil {
		uid := *entry.UserID
		userID = &uid
	}

	log := &model.AuditLog{
		LogID:          generateUUID(),
		UserID:         userID,
		EventTime:      eventTime,
		ActionType:     entry.ActionType,
		TargetResource: entry.TargetResource,
		SourceIP:       entry.SourceIP,
		Result:         entry.Result,
		Details:        entry.Details,
		CreatedAt:      time.Now(),
	}

	detailsJSON := "null"
	if entry.Details != nil {
		detailsJSON = string(entry.Details)
	}

	hashInput := fmt.Sprintf("%v|%s|%s|%s|%s|%s|%v",
		userID,
		entry.ActionType,
		entry.TargetResource,
		entry.SourceIP,
		entry.Result,
		detailsJSON,
		eventTime.UnixNano(),
	)
	log.SM3Hash = computeSM3Hash(hashInput)

	if err := s.repo.Create(ctx, log); err != nil {
		return nil, err
	}

	return log, nil
}

func (s *auditService) RecordBatch(ctx context.Context, entries []model.AuditLogEntry) (*model.BatchAuditLogResponse, error) {
	response := &model.BatchAuditLogResponse{
		Logs: make([]model.AuditLogResponse, 0, len(entries)),
	}

	logs := make([]*model.AuditLog, 0, len(entries))

	for _, entry := range entries {
		eventTime := time.Now()
		if entry.EventTime != nil {
			eventTime = *entry.EventTime
		}

		var userID *int64
		if entry.UserID != nil {
			uid := *entry.UserID
			userID = &uid
		}

		log := &model.AuditLog{
			LogID:          generateUUID(),
			UserID:         userID,
			EventTime:      eventTime,
			ActionType:     entry.ActionType,
			TargetResource: entry.TargetResource,
			SourceIP:       entry.SourceIP,
			Result:         entry.Result,
			Details:        entry.Details,
			CreatedAt:      time.Now(),
		}

		detailsJSON := "null"
		if entry.Details != nil {
			detailsJSON = string(entry.Details)
		}

		hashInput := fmt.Sprintf("%v|%s|%s|%s|%s|%s|%v",
			userID,
			entry.ActionType,
			entry.TargetResource,
			entry.SourceIP,
			entry.Result,
			detailsJSON,
			eventTime.UnixNano(),
		)
		log.SM3Hash = computeSM3Hash(hashInput)

		logs = append(logs, log)
		response.Logs = append(response.Logs, model.AuditLogResponse{
			LogID:   log.LogID,
			SM3Hash: log.SM3Hash,
		})
	}

	if err := s.repo.CreateBatch(ctx, logs); err != nil {
		response.Failed = len(entries)
		response.Processed = 0
		return response, err
	}

	response.Processed = len(entries)
	response.Failed = 0
	return response, nil
}

func (s *auditService) GetLogsByUser(ctx context.Context, userID string, limit, offset int) ([]*model.AuditLog, error) {
	if limit <= 0 {
		limit = s.cfg.AuditLogDefaultLimit
	}
	if offset < 0 {
		offset = 0
	}

	var uid int64
	_, err := fmt.Sscanf(userID, "%d", &uid)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	return s.repo.GetByUserID(ctx, uid, limit, offset)
}

func (s *auditService) GetLogsByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*model.AuditLog, error) {
	if limit <= 0 {
		limit = s.cfg.AuditLogDefaultLimit
	}
	if offset < 0 {
		offset = 0
	}

	return s.repo.GetByTimeRange(ctx, start, end, limit, offset)
}

func (s *auditService) VerifyLogIntegrity(ctx context.Context, logID string) (*model.IntegrityVerifyResponse, error) {
	log, err := s.repo.GetByLogID(ctx, logID)
	if err != nil {
		return nil, err
	}
	if log == nil {
		return nil, ErrLogNotFound
	}

	detailsJSON := "null"
	if log.Details != nil {
		detailsJSON = string(log.Details)
	}

	hashInput := fmt.Sprintf("%v|%s|%s|%s|%s|%s|%v",
		log.UserID,
		log.ActionType,
		log.TargetResource,
		log.SourceIP,
		log.Result,
		detailsJSON,
		log.EventTime.UnixNano(),
	)
	computedHash := computeSM3Hash(hashInput)

	return &model.IntegrityVerifyResponse{
		LogID:         logID,
		IsValid:       computedHash == log.SM3Hash,
		StoredHash:    log.SM3Hash,
		ComputedHash:  computedHash,
	}, nil
}

func (s *auditService) CleanupOldLogs(ctx context.Context, retentionDays int) (*model.CleanupResponse, error) {
	if retentionDays <= 0 {
		return nil, ErrInvalidRetention
	}

	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	deletedCount, err := s.repo.DeleteOlderThan(ctx, cutoffTime)
	if err != nil {
		return nil, err
	}

	return &model.CleanupResponse{
		DeletedCount:  deletedCount,
		RetentionDays: retentionDays,
	}, nil
}

func computeSM3Hash(data string) string {
	h := crypto.NewSM3()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func generateUUID() string {
	return uuid.New().String()
}
