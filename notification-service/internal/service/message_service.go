package service

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
)

type MessageRepository interface {
	Create(ctx context.Context, msg *model.Message) error
	ListByUserID(ctx context.Context, userID int64, offset, limit int) ([]*model.Message, error)
	MarkRead(ctx context.Context, id int64) error
	MarkAllRead(ctx context.Context, userID int64) error
	CountUnread(ctx context.Context, userID int64) (int64, error)
}

type MessageService struct {
	repo      MessageRepository
	messages  []*model.Message
	mu        sync.RWMutex
	idCounter atomic.Int64
}

func NewMessageService(repo MessageRepository) *MessageService {
	return &MessageService{repo: repo}
}

func (s *MessageService) CreateMessage(ctx context.Context, msg *model.Message) (*model.Message, error) {
	msg.ID = s.idCounter.Add(1)
	msg.CreatedAt = time.Now()
	msg.Read = false

	if s.repo != nil {
		if err := s.repo.Create(ctx, msg); err != nil {
			return nil, fmt.Errorf("failed to create message: %w", err)
		}
		return msg, nil
	}

	s.mu.Lock()
	s.messages = append(s.messages, msg)
	s.mu.Unlock()
	return msg, nil
}

func (s *MessageService) ListMessages(ctx context.Context, userID int64, offset, limit int) ([]*model.Message, error) {
	if s.repo != nil {
		return s.repo.ListByUserID(ctx, userID, offset, limit)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*model.Message
	for _, m := range s.messages {
		if m.UserID == userID {
			result = append(result, m)
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

func (s *MessageService) MarkRead(ctx context.Context, id int64) error {
	if s.repo != nil {
		return s.repo.MarkRead(ctx, id)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, m := range s.messages {
		if m.ID == id {
			m.Read = true
			return nil
		}
	}
	return fmt.Errorf("message not found")
}

func (s *MessageService) MarkAllRead(ctx context.Context, userID int64) error {
	if s.repo != nil {
		return s.repo.MarkAllRead(ctx, userID)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, m := range s.messages {
		if m.UserID == userID {
			m.Read = true
		}
	}
	return nil
}

func (s *MessageService) GetUnreadCount(ctx context.Context, userID int64) (int64, error) {
	if s.repo != nil {
		return s.repo.CountUnread(ctx, userID)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int64
	for _, m := range s.messages {
		if m.UserID == userID && !m.Read {
			count++
		}
	}
	return count, nil
}
