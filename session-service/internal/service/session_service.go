package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/sunxi/92-Account-Center/session-service/internal/model"
	"github.com/sunxi/92-Account-Center/session-service/internal/repository"
)

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionExpired      = errors.New("session expired")
	ErrMaxSessionsReached  = errors.New("maximum concurrent sessions reached")
)

const (
	DefaultMaxSessions = 5
	SessionTimeout     = 20 * time.Minute
)

type SessionService interface {
	CreateSession(ctx context.Context, req *model.CreateSessionRequest) (*model.Session, error)
	ValidateSession(ctx context.Context, sessionID string) (*model.SessionInfo, error)
	GetUserSessions(ctx context.Context, userID int64) ([]*model.SessionInfo, error)
	InvalidateSession(ctx context.Context, sessionID string) error
	InvalidateAllUserSessions(ctx context.Context, userID int64) error
	RefreshSession(ctx context.Context, sessionID string) (*model.SessionInfo, error)
	CountUserSessions(ctx context.Context, userID int64) (int64, error)
}

type sessionService struct {
	repo           repository.SessionRepository
	maxConSessions int64
}

func NewSessionService(repo repository.SessionRepository, maxConSessions int64) SessionService {
	if maxConSessions <= 0 {
		maxConSessions = DefaultMaxSessions
	}
	return &sessionService{
		repo:           repo,
		maxConSessions: maxConSessions,
	}
}

func (s *sessionService) CreateSession(ctx context.Context, req *model.CreateSessionRequest) (*model.Session, error) {
	count, err := s.repo.CountUserSessions(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	if count >= s.maxConSessions {
		sessions, err := s.repo.GetUserSessions(ctx, req.UserID)
		if err != nil {
			return nil, err
		}

		if len(sessions) > 0 {
			oldest := sessions[0]
			for _, session := range sessions {
				if session.CreatedAt.Before(oldest.CreatedAt) {
					oldest = session
				}
			}
			s.repo.Delete(ctx, oldest.SessionID, oldest.UserID)
		}
	}

	now := time.Now()
	session := &model.Session{
		SessionID:         uuid.New().String(),
		UserID:            req.UserID,
		DeviceFingerprint: req.DeviceFingerprint,
		IPAddress:         req.IPAddress,
		CreatedAt:         now,
		LastAccessedAt:    now,
		ExpiresAt:         now.Add(SessionTimeout),
		IsActive:          true,
	}

	if err := s.repo.Create(ctx, session); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *sessionService) ValidateSession(ctx context.Context, sessionID string) (*model.SessionInfo, error) {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	if !session.IsActive {
		return nil, ErrSessionExpired
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	ttl, err := s.repo.GetSessionTTL(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if ttl > 0 {
		newExpiresAt := time.Now().Add(ttl)
		s.repo.UpdateLastAccessed(ctx, sessionID, time.Now(), newExpiresAt)
		session.ExpiresAt = newExpiresAt
	}

	return session.ToSessionInfo(int64(ttl.Seconds())), nil
}

func (s *sessionService) GetUserSessions(ctx context.Context, userID int64) ([]*model.SessionInfo, error) {
	sessions, err := s.repo.GetUserSessions(ctx, userID)
	if err != nil {
		return nil, err
	}

	var sessionInfos []*model.SessionInfo
	for _, session := range sessions {
		ttl, err := s.repo.GetSessionTTL(ctx, session.SessionID)
		if err != nil {
			continue
		}
		sessionInfos = append(sessionInfos, session.ToSessionInfo(int64(ttl.Seconds())))
	}

	return sessionInfos, nil
}

func (s *sessionService) InvalidateSession(ctx context.Context, sessionID string) error {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return ErrSessionNotFound
		}
		return err
	}

	return s.repo.Delete(ctx, sessionID, session.UserID)
}

func (s *sessionService) InvalidateAllUserSessions(ctx context.Context, userID int64) error {
	return s.repo.DeleteAllUserSessions(ctx, userID)
}

func (s *sessionService) RefreshSession(ctx context.Context, sessionID string) (*model.SessionInfo, error) {
	session, err := s.repo.GetByID(ctx, sessionID)
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	if !session.IsActive {
		return nil, ErrSessionExpired
	}

	now := time.Now()
	newExpiresAt := now.Add(SessionTimeout)

	if err := s.repo.UpdateLastAccessed(ctx, sessionID, now, newExpiresAt); err != nil {
		return nil, err
	}

	session.LastAccessedAt = now
	session.ExpiresAt = newExpiresAt

	return session.ToSessionInfo(int64(SessionTimeout.Seconds())), nil
}

func (s *sessionService) CountUserSessions(ctx context.Context, userID int64) (int64, error) {
	return s.repo.CountUserSessions(ctx, userID)
}
