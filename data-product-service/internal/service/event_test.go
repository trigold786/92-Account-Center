package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
)

func TestEventValidation(t *testing.T) {
	svc := NewEventService(nil)
	err := svc.ValidateEvent(context.Background(), "page_view", map[string]interface{}{"url": "/home"})
	if err != nil {
		t.Fatalf("ValidateEvent failed: %v", err)
	}
	err = svc.ValidateEvent(context.Background(), "", nil)
	if err == nil {
		t.Fatal("expected error for empty event type")
	}
}

func TestBatchEventProcessing(t *testing.T) {
	repo := &mockEventRepo{}
	svc := NewEventService(repo)
	events := []model.Event{
		{EventType: "page_view", UserID: 1, Properties: map[string]interface{}{"url": "/home"}},
		{EventType: "click", UserID: 1, Properties: map[string]interface{}{"element": "signup_btn"}},
	}
	err := svc.BatchProcess(context.Background(), events)
	if err != nil {
		t.Fatalf("BatchProcess failed: %v", err)
	}
}

type mockEventRepo struct{}

func (m *mockEventRepo) BatchInsert(ctx context.Context, events []model.Event) error {
	return nil
}
