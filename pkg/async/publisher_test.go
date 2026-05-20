package async

import (
	"context"
	"testing"
)

func TestPublish(t *testing.T) {
	pub := NewPublisher(nil, "test-stream")
	if pub.Stream != "test-stream" {
		t.Fatalf("unexpected stream: %s", pub.Stream)
	}
	err := pub.Publish(context.Background(), "event_type", map[string]interface{}{"key": "value"}, "trace_123")
	if err != nil {
		t.Fatalf("Publish failed (expected with nil redis): %v", err)
	}
}

func TestPublishWithHeaders(t *testing.T) {
	pub := NewPublisher(nil, "events")
	msg := NewMessage("user.created", map[string]interface{}{"user_id": 1})
	msg.TraceID = "trace_abc"
	msg.SetHeader("X-Source", "auth-service")
	if msg.EventType != "user.created" {
		t.Fatalf("unexpected event type: %s", msg.EventType)
	}
	err := pub.PublishMessage(context.Background(), msg)
	if err != nil {
		t.Fatalf("PublishMessage failed: %v", err)
	}
}
