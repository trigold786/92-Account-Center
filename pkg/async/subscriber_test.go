package async

import (
	"context"
	"sync/atomic"
	"testing"
)

func TestSubscribe(t *testing.T) {
	var count atomic.Int32
	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		count.Add(1)
		return nil
	})
	sub := NewSubscriber(nil, "events", "group-1", "consumer-1")
	sub.RegisterHandler("user.created", handler)
	if handlers, ok := sub.handlers["user.created"]; !ok || len(handlers) == 0 {
		t.Fatal("handler should be registered")
	}
}
