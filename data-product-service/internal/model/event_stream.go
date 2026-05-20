package model

import "time"

type StreamEvent struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	EventType string    `json:"event_type"`
	Payload   string    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}
