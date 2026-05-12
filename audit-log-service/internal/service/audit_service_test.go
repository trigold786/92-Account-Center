package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/audit-log-service/internal/model"
)

type mockAuditRepo struct {
	logs     map[string]*model.AuditLog
	createErr error
}

func newMockAuditRepo() *mockAuditRepo {
	return &mockAuditRepo{
		logs: make(map[string]*model.AuditLog),
	}
}

func (m *mockAuditRepo) Create(ctx context.Context, log *model.AuditLog) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.logs[log.LogID] = log
	return nil
}

func (m *mockAuditRepo) CreateBatch(ctx context.Context, logs []*model.AuditLog) error {
	if m.createErr != nil {
		return m.createErr
	}
	for _, l := range logs {
		m.logs[l.LogID] = l
	}
	return nil
}

func (m *mockAuditRepo) GetByLogID(ctx context.Context, logID string) (*model.AuditLog, error) {
	l, ok := m.logs[logID]
	if !ok {
		return nil, nil
	}
	return l, nil
}

func (m *mockAuditRepo) GetByUserID(ctx context.Context, userID int64, limit, offset int) ([]*model.AuditLog, error) {
	var result []*model.AuditLog
	for _, l := range m.logs {
		if l.UserID != nil && *l.UserID == userID {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockAuditRepo) GetByTimeRange(ctx context.Context, start, end time.Time, limit, offset int) ([]*model.AuditLog, error) {
	var result []*model.AuditLog
	for _, l := range m.logs {
		if !l.EventTime.Before(start) && !l.EventTime.After(end) {
			result = append(result, l)
		}
	}
	return result, nil
}

func (m *mockAuditRepo) DeleteOlderThan(ctx context.Context, cutoffTime time.Time) (int64, error) {
	var count int64
	for id, l := range m.logs {
		if l.EventTime.Before(cutoffTime) {
			delete(m.logs, id)
			count++
		}
	}
	return count, nil
}

func TestRecordLog(t *testing.T) {
	repo := newMockAuditRepo()
	svc := NewAuditService(repo)

	userID := int64(1)
	entry := &model.AuditLogEntry{
		UserID:         &userID,
		ActionType:     "LOGIN",
		TargetResource: "/auth/login",
		SourceIP:       "127.0.0.1",
		Result:         "success",
		Details:        json.RawMessage(`{"browser":"chrome"}`),
	}

	log, err := svc.RecordLog(context.Background(), entry)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if log.LogID == "" {
		t.Error("expected non-empty log ID")
	}
	if log.SM3Hash == "" {
		t.Error("expected non-empty SM3 hash")
	}
	if log.ActionType != "LOGIN" {
		t.Errorf("expected action type LOGIN, got %s", log.ActionType)
	}
	if log.Result != "success" {
		t.Errorf("expected result success, got %s", log.Result)
	}
}

func TestRecordLog_RepoError(t *testing.T) {
	repo := newMockAuditRepo()
	repo.createErr = errors.New("db error")
	svc := NewAuditService(repo)

	userID := int64(1)
	entry := &model.AuditLogEntry{
		UserID:         &userID,
		ActionType:     "LOGIN",
		TargetResource: "/auth/login",
		SourceIP:       "127.0.0.1",
		Result:         "success",
	}

	_, err := svc.RecordLog(context.Background(), entry)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRecordBatch(t *testing.T) {
	repo := newMockAuditRepo()
	svc := NewAuditService(repo)

	userID := int64(1)
	entries := []model.AuditLogEntry{
		{UserID: &userID, ActionType: "LOGIN", TargetResource: "/auth/login", SourceIP: "127.0.0.1", Result: "success"},
		{UserID: &userID, ActionType: "LOGOUT", TargetResource: "/auth/logout", SourceIP: "127.0.0.1", Result: "success"},
	}

	resp, err := svc.RecordBatch(context.Background(), entries)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.Processed != 2 {
		t.Errorf("expected 2 processed, got %d", resp.Processed)
	}
	if resp.Failed != 0 {
		t.Errorf("expected 0 failed, got %d", resp.Failed)
	}
	if len(resp.Logs) != 2 {
		t.Errorf("expected 2 logs, got %d", len(resp.Logs))
	}
	for _, l := range resp.Logs {
		if l.SM3Hash == "" {
			t.Error("expected non-empty SM3 hash in batch response")
		}
	}
}

func TestRecordBatch_RepoError(t *testing.T) {
	repo := newMockAuditRepo()
	repo.createErr = errors.New("db error")
	svc := NewAuditService(repo)

	userID := int64(1)
	entries := []model.AuditLogEntry{
		{UserID: &userID, ActionType: "LOGIN", TargetResource: "/auth/login", SourceIP: "127.0.0.1", Result: "success"},
	}

	resp, err := svc.RecordBatch(context.Background(), entries)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if resp.Processed != 0 {
		t.Errorf("expected 0 processed, got %d", resp.Processed)
	}
	if resp.Failed != 1 {
		t.Errorf("expected 1 failed, got %d", resp.Failed)
	}
}

func TestGetLogsByUser(t *testing.T) {
	repo := newMockAuditRepo()
	svc := NewAuditService(repo)

	userID := int64(1)
	entry := &model.AuditLogEntry{
		UserID:         &userID,
		ActionType:     "LOGIN",
		TargetResource: "/auth/login",
		SourceIP:       "127.0.0.1",
		Result:         "success",
	}
	svc.RecordLog(context.Background(), entry)

	logs, err := svc.GetLogsByUser(context.Background(), "1", 10, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
}

func TestGetLogsByUser_InvalidUserID(t *testing.T) {
	repo := newMockAuditRepo()
	svc := NewAuditService(repo)

	_, err := svc.GetLogsByUser(context.Background(), "abc", 10, 0)
	if err == nil {
		t.Fatal("expected error for invalid user ID, got nil")
	}
}

func TestGetLogsByUser_DefaultLimit(t *testing.T) {
	repo := newMockAuditRepo()
	svc := NewAuditService(repo)

	userID := int64(1)
	svc.RecordLog(context.Background(), &model.AuditLogEntry{
		UserID: &userID, ActionType: "LOGIN", TargetResource: "/a", SourceIP: "1.1.1.1", Result: "success",
	})

	logs, err := svc.GetLogsByUser(context.Background(), "1", 0, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(logs) != 1 {
		t.Errorf("expected 1 log, got %d", len(logs))
	}
}

func TestVerifyLogIntegrity_Valid(t *testing.T) {
	repo := newMockAuditRepo()
	svc := NewAuditService(repo)

	userID := int64(1)
	entry := &model.AuditLogEntry{
		UserID:         &userID,
		ActionType:     "LOGIN",
		TargetResource: "/auth/login",
		SourceIP:       "127.0.0.1",
		Result:         "success",
		Details:        json.RawMessage(`{"key":"value"}`),
	}
	log, _ := svc.RecordLog(context.Background(), entry)

	resp, err := svc.VerifyLogIntegrity(context.Background(), log.LogID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !resp.IsValid {
		t.Errorf("expected valid integrity, got stored=%s computed=%s", resp.StoredHash, resp.ComputedHash)
	}
}

func TestVerifyLogIntegrity_Invalid(t *testing.T) {
	repo := newMockAuditRepo()
	svc := NewAuditService(repo)

	userID := int64(1)
	entry := &model.AuditLogEntry{
		UserID:         &userID,
		ActionType:     "LOGIN",
		TargetResource: "/auth/login",
		SourceIP:       "127.0.0.1",
		Result:         "success",
	}
	log, _ := svc.RecordLog(context.Background(), entry)

	stored := repo.logs[log.LogID]
	stored.SM3Hash = "tampered_hash"

	resp, err := svc.VerifyLogIntegrity(context.Background(), log.LogID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.IsValid {
		t.Error("expected invalid integrity for tampered hash")
	}
}

func TestVerifyLogIntegrity_NotFound(t *testing.T) {
	repo := newMockAuditRepo()
	svc := NewAuditService(repo)

	_, err := svc.VerifyLogIntegrity(context.Background(), "nonexistent")
	if err != ErrLogNotFound {
		t.Errorf("expected ErrLogNotFound, got %v", err)
	}
}

func TestCleanupOldLogs(t *testing.T) {
	repo := newMockAuditRepo()
	svc := NewAuditService(repo)

	userID := int64(1)
	oldTime := time.Now().Add(-100 * 24 * time.Hour)
	oldEntry := &model.AuditLogEntry{
		UserID:         &userID,
		ActionType:     "LOGIN",
		TargetResource: "/auth/login",
		SourceIP:       "127.0.0.1",
		Result:         "success",
		EventTime:      &oldTime,
	}
	svc.RecordLog(context.Background(), oldEntry)

	resp, err := svc.CleanupOldLogs(context.Background(), 30)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.DeletedCount != 1 {
		t.Errorf("expected 1 deleted, got %d", resp.DeletedCount)
	}
	if resp.RetentionDays != 30 {
		t.Errorf("expected retention days 30, got %d", resp.RetentionDays)
	}
}

func TestCleanupOldLogs_InvalidRetention(t *testing.T) {
	repo := newMockAuditRepo()
	svc := NewAuditService(repo)

	_, err := svc.CleanupOldLogs(context.Background(), 0)
	if err != ErrInvalidRetention {
		t.Errorf("expected ErrInvalidRetention, got %v", err)
	}

	_, err = svc.CleanupOldLogs(context.Background(), -1)
	if err != ErrInvalidRetention {
		t.Errorf("expected ErrInvalidRetention, got %v", err)
	}
}
