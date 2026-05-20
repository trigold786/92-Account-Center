package service

import (
	"context"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
)

func TestProcessEvent(t *testing.T) {
	svc := NewStreamService(nil)

	e, err := svc.ProcessEvent(context.Background(), &model.StreamEvent{
		UserID:    1,
		EventType: "page_view",
		Payload:   `{"page":"/home"}`,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if e.Timestamp.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}

func TestOnlineCount(t *testing.T) {
	svc := NewStreamService(nil)

	svc.ProcessEvent(context.Background(), &model.StreamEvent{UserID: 1, EventType: "page_view"})
	svc.ProcessEvent(context.Background(), &model.StreamEvent{UserID: 2, EventType: "page_view"})
	svc.ProcessEvent(context.Background(), &model.StreamEvent{UserID: 1, EventType: "click"})

	count, err := svc.GetOnlineCount(context.Background(), 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 unique users, got %d", count)
	}
}

func TestRealtimeFunnel(t *testing.T) {
	svc := NewStreamService(nil)

	svc.ProcessEvent(context.Background(), &model.StreamEvent{UserID: 1, EventType: "register"})
	svc.ProcessEvent(context.Background(), &model.StreamEvent{UserID: 1, EventType: "subscribe"})
	svc.ProcessEvent(context.Background(), &model.StreamEvent{UserID: 2, EventType: "register"})
	svc.ProcessEvent(context.Background(), &model.StreamEvent{UserID: 3, EventType: "register"})
	svc.ProcessEvent(context.Background(), &model.StreamEvent{UserID: 3, EventType: "subscribe"})
	svc.ProcessEvent(context.Background(), &model.StreamEvent{UserID: 3, EventType: "pay"})

	steps := []string{"register", "subscribe", "pay"}
	funnel, err := svc.GetRealtimeFunnel(context.Background(), 1*time.Minute, steps)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(funnel) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(funnel))
	}
	if funnel[0].Count != 3 {
		t.Fatalf("expected 3 registers, got %d", funnel[0].Count)
	}
	if funnel[1].Count != 2 {
		t.Fatalf("expected 2 subscribes, got %d", funnel[1].Count)
	}
	if funnel[2].Count != 1 {
		t.Fatalf("expected 1 pay, got %d", funnel[2].Count)
	}
}
