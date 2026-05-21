package service

import (
	"context"
	"errors"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
)

var validEventTypes = map[string]bool{
	"page_view": true, "click": true, "session_start": true, "session_end": true,
	"login": true, "register": true, "subscribe": true, "upgrade": true, "downgrade": true,
	"payment_start": true, "payment_success": true, "payment_fail": true,
	"referral_share": true, "referral_register": true, "ad_shown": true,
}

type EventRepository interface {
	BatchInsert(ctx context.Context, events []model.Event) error
}

type EventService struct {
	repo EventRepository
}

func NewEventService(repo EventRepository) *EventService {
	return &EventService{repo: repo}
}

func (s *EventService) ValidateEvent(ctx context.Context, eventType string, properties map[string]interface{}) error {
	if eventType == "" {
		return errors.New("event_type is required")
	}
	if !validEventTypes[eventType] {
		return errors.New("invalid event_type: " + eventType)
	}
	return nil
}

func (s *EventService) BatchProcess(ctx context.Context, events []model.Event) error {
	for _, e := range events {
		if err := s.ValidateEvent(ctx, e.EventType, e.Properties); err != nil {
			return err
		}
	}
	return s.repo.BatchInsert(ctx, events)
}
