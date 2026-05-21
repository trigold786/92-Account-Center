package model

import "time"

type AdEvent struct {
	ID         int64     `json:"id"`
	UserID     int64     `json:"user_id"`
	EventType  string    `json:"event_type"`
	AdID       string    `json:"ad_id"`
	Placement  string    `json:"placement"`
	Duration   int       `json:"duration,omitempty"`
	Converted  bool      `json:"converted"`
	Timestamp  time.Time `json:"timestamp"`
}
