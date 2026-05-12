package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/session-service/internal/model"
	"github.com/trigold786/92-Account-Center/session-service/internal/repository"
)

type mockSessionRepo struct {
	sessions     map[string]*model.Session
	userSessions map[int64][]string
	createErr    error
	ttl          time.Duration
	ttlErr       error
}

func newMockSessionRepo() *mockSessionRepo {
	return &mockSessionRepo{
		sessions:     make(map[string]*model.Session),
		userSessions: make(map[int64][]string),
		ttl:          15 * time.Minute,
	}
}

func (m *mockSessionRepo) Create(ctx context.Context, session *model.Session) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.sessions[session.SessionID] = session
	m.userSessions[session.UserID] = append(m.userSessions[session.UserID], session.SessionID)
	return nil
}

func (m *mockSessionRepo) CreateWithEviction(ctx context.Context, session *model.Session, maxSessions int64) error {
	if m.createErr != nil {
		return m.createErr
	}
	ids := m.userSessions[session.UserID]
	var valid []string
	for _, id := range ids {
		if _, ok := m.sessions[id]; ok {
			valid = append(valid, id)
		}
	}
	if int64(len(valid)) >= maxSessions {
		evictID := valid[0]
		delete(m.sessions, evictID)
		var remaining []string
		for _, id := range valid[1:] {
			remaining = append(remaining, id)
		}
		m.userSessions[session.UserID] = remaining
	}
	m.sessions[session.SessionID] = session
	m.userSessions[session.UserID] = append(m.userSessions[session.UserID], session.SessionID)
	return nil
}

func (m *mockSessionRepo) GetByID(ctx context.Context, sessionID string) (*model.Session, error) {
	s, ok := m.sessions[sessionID]
	if !ok {
		return nil, repository.ErrSessionNotFound
	}
	return s, nil
}

func (m *mockSessionRepo) GetUserSessions(ctx context.Context, userID int64) ([]*model.Session, error) {
	ids := m.userSessions[userID]
	var result []*model.Session
	for _, id := range ids {
		if s, ok := m.sessions[id]; ok && s.IsActive {
			result = append(result, s)
		}
	}
	return result, nil
}

func (m *mockSessionRepo) UpdateLastAccessed(ctx context.Context, sessionID string, lastAccessedAt time.Time, expiresAt time.Time) error {
	s, ok := m.sessions[sessionID]
	if !ok {
		return repository.ErrSessionNotFound
	}
	s.LastAccessedAt = lastAccessedAt
	s.ExpiresAt = expiresAt
	return nil
}

func (m *mockSessionRepo) Delete(ctx context.Context, sessionID string, userID int64) error {
	delete(m.sessions, sessionID)
	ids := m.userSessions[userID]
	var remaining []string
	for _, id := range ids {
		if id != sessionID {
			remaining = append(remaining, id)
		}
	}
	m.userSessions[userID] = remaining
	return nil
}

func (m *mockSessionRepo) DeleteAllUserSessions(ctx context.Context, userID int64) error {
	ids := m.userSessions[userID]
	for _, id := range ids {
		delete(m.sessions, id)
	}
	delete(m.userSessions, userID)
	return nil
}

func (m *mockSessionRepo) CountUserSessions(ctx context.Context, userID int64) (int64, error) {
	ids := m.userSessions[userID]
	var count int64
	for _, id := range ids {
		if _, ok := m.sessions[id]; ok {
			count++
		}
	}
	return count, nil
}

func (m *mockSessionRepo) GetSessionTTL(ctx context.Context, sessionID string) (time.Duration, error) {
	if m.ttlErr != nil {
		return 0, m.ttlErr
	}
	return m.ttl, nil
}

func TestCreateSession(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, 5)

	req := &model.CreateSessionRequest{
		UserID:            1,
		DeviceFingerprint: "fp-abc",
		IPAddress:         "127.0.0.1",
	}

	session, err := svc.CreateSession(context.Background(), req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session.SessionID == "" {
		t.Error("expected non-empty session ID")
	}
	if session.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", session.UserID)
	}
	if !session.IsActive {
		t.Error("expected session to be active")
	}
	if session.ExpiresAt.Before(time.Now()) {
		t.Error("expected expires_at to be in the future")
	}
}

func TestCreateSession_RepoError(t *testing.T) {
	repo := newMockSessionRepo()
	repo.createErr = errors.New("db error")
	svc := NewSessionService(repo, 5)

	req := &model.CreateSessionRequest{UserID: 1, DeviceFingerprint: "fp", IPAddress: "1.1.1.1"}
	_, err := svc.CreateSession(context.Background(), req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestValidateSession_Active(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, 5)

	req := &model.CreateSessionRequest{UserID: 1, DeviceFingerprint: "fp", IPAddress: "1.1.1.1"}
	session, _ := svc.CreateSession(context.Background(), req)

	info, err := svc.ValidateSession(context.Background(), session.SessionID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.SessionID != session.SessionID {
		t.Errorf("expected session ID %s, got %s", session.SessionID, info.SessionID)
	}
	if info.RemainingTTL <= 0 {
		t.Error("expected positive remaining TTL")
	}
}

func TestValidateSession_Expired(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, 5)

	req := &model.CreateSessionRequest{UserID: 1, DeviceFingerprint: "fp", IPAddress: "1.1.1.1"}
	session, _ := svc.CreateSession(context.Background(), req)

	session.ExpiresAt = time.Now().Add(-1 * time.Hour)

	_, err := svc.ValidateSession(context.Background(), session.SessionID)
	if err != ErrSessionExpired {
		t.Errorf("expected ErrSessionExpired, got %v", err)
	}
}

func TestValidateSession_NotFound(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, 5)

	_, err := svc.ValidateSession(context.Background(), "nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestInvalidateSession(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, 5)

	req := &model.CreateSessionRequest{UserID: 1, DeviceFingerprint: "fp", IPAddress: "1.1.1.1"}
	session, _ := svc.CreateSession(context.Background(), req)

	err := svc.InvalidateSession(context.Background(), session.SessionID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = svc.ValidateSession(context.Background(), session.SessionID)
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound after invalidation, got %v", err)
	}
}

func TestInvalidateSession_NotFound(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, 5)

	err := svc.InvalidateSession(context.Background(), "nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestInvalidateAllUserSessions(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, 5)

	for i := 0; i < 3; i++ {
		req := &model.CreateSessionRequest{UserID: 1, DeviceFingerprint: "fp", IPAddress: "1.1.1.1"}
		svc.CreateSession(context.Background(), req)
	}

	err := svc.InvalidateAllUserSessions(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	count, _ := svc.CountUserSessions(context.Background(), 1)
	if count != 0 {
		t.Errorf("expected 0 sessions after invalidation, got %d", count)
	}
}

func TestRefreshSession(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, 5)

	req := &model.CreateSessionRequest{UserID: 1, DeviceFingerprint: "fp", IPAddress: "1.1.1.1"}
	session, _ := svc.CreateSession(context.Background(), req)

	info, err := svc.RefreshSession(context.Background(), session.SessionID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if info.SessionID != session.SessionID {
		t.Errorf("expected session ID %s, got %s", session.SessionID, info.SessionID)
	}
	if info.ExpiresAt.Before(time.Now()) {
		t.Error("expected refreshed expires_at to be in the future")
	}
}

func TestRefreshSession_NotFound(t *testing.T) {
	repo := newMockSessionRepo()
	svc := NewSessionService(repo, 5)

	_, err := svc.RefreshSession(context.Background(), "nonexistent")
	if err != ErrSessionNotFound {
		t.Errorf("expected ErrSessionNotFound, got %v", err)
	}
}

func TestMaxConcurrentSessions(t *testing.T) {
	repo := newMockSessionRepo()
	maxSessions := int64(2)
	svc := NewSessionService(repo, maxSessions)

	s1, _ := svc.CreateSession(context.Background(), &model.CreateSessionRequest{UserID: 1, DeviceFingerprint: "fp1", IPAddress: "1.1.1.1"})
	svc.CreateSession(context.Background(), &model.CreateSessionRequest{UserID: 1, DeviceFingerprint: "fp2", IPAddress: "1.1.1.2"})
	svc.CreateSession(context.Background(), &model.CreateSessionRequest{UserID: 1, DeviceFingerprint: "fp3", IPAddress: "1.1.1.3"})

	count, err := svc.CountUserSessions(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count > maxSessions {
		t.Errorf("expected at most %d sessions, got %d", maxSessions, count)
	}

	_, err = repo.GetByID(context.Background(), s1.SessionID)
	if err != repository.ErrSessionNotFound {
		t.Errorf("expected oldest session to be evicted, got err=%v", err)
	}
}
