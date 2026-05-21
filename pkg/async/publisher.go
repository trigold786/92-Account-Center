package async

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type Message struct {
	ID         string                 `json:"id"`
	EventType  string                 `json:"event_type"`
	Data       map[string]interface{} `json:"data"`
	Headers    map[string]string      `json:"headers"`
	TraceID    string                 `json:"trace_id"`
	Timestamp  time.Time              `json:"timestamp"`
	RetryCount int                    `json:"retry_count"`
}

func NewMessage(eventType string, data map[string]interface{}) *Message {
	return &Message{
		EventType: eventType,
		Data:      data,
		Headers:   make(map[string]string),
		Timestamp: time.Now(),
	}
}

func (m *Message) SetHeader(key, value string) {
	m.Headers[key] = value
}

func (m *Message) GetHeader(key string) string {
	return m.Headers[key]
}

type Publisher struct {
	rdb    interface{}
	Stream string
}

func NewPublisher(rdb interface{}, stream string) *Publisher {
	return &Publisher{rdb: rdb, Stream: stream}
}

func (p *Publisher) Publish(ctx context.Context, eventType string, data map[string]interface{}, traceID string) error {
	msg := NewMessage(eventType, data)
	msg.TraceID = traceID
	return p.PublishMessage(ctx, msg)
}

func (p *Publisher) PublishMessage(ctx context.Context, msg *Message) error {
	payload, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	if p.rdb == nil {
		fmt.Printf("[async] publish to %s: %s\n", p.Stream, string(payload))
		return nil
	}
	return nil
}
