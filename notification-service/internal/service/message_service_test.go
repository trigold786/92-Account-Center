package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
)

func TestCreateMessage(t *testing.T) {
	svc := NewMessageService(nil)

	msg, err := svc.CreateMessage(context.Background(), &model.Message{
		UserID: 1,
		Type:   "system",
		Title:  "Welcome",
		Body:   "Welcome to our platform",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
	if msg.Read {
		t.Fatal("expected unread")
	}
}

func TestListMessages(t *testing.T) {
	svc := NewMessageService(nil)

	svc.CreateMessage(context.Background(), &model.Message{UserID: 1, Type: "system", Title: "Msg1"})
	svc.CreateMessage(context.Background(), &model.Message{UserID: 1, Type: "promo", Title: "Msg2"})
	svc.CreateMessage(context.Background(), &model.Message{UserID: 2, Type: "system", Title: "Msg3"})

	msgs, err := svc.ListMessages(context.Background(), 1, 0, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
}

func TestMarkRead(t *testing.T) {
	svc := NewMessageService(nil)

	msg, _ := svc.CreateMessage(context.Background(), &model.Message{UserID: 1, Type: "system", Title: "Test"})

	err := svc.MarkRead(context.Background(), msg.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	msgs, _ := svc.ListMessages(context.Background(), 1, 0, 10)
	if !msgs[0].Read {
		t.Fatal("expected message to be read")
	}
}

func TestUnreadCount(t *testing.T) {
	svc := NewMessageService(nil)

	svc.CreateMessage(context.Background(), &model.Message{UserID: 1, Type: "system", Title: "Msg1"})
	svc.CreateMessage(context.Background(), &model.Message{UserID: 1, Type: "system", Title: "Msg2"})
	svc.CreateMessage(context.Background(), &model.Message{UserID: 1, Type: "system", Title: "Msg3"})

	count, err := svc.GetUnreadCount(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3 unread, got %d", count)
	}

	svc.MarkRead(context.Background(), 1)
	count, _ = svc.GetUnreadCount(context.Background(), 1)
	if count != 2 {
		t.Fatalf("expected 2 unread after mark read, got %d", count)
	}
}
