package model

import "time"

type Event struct {
	ID         int64                  `json:"id"`
	EventType  string                 `json:"event_type"`
	UserID     int64                  `json:"user_id"`
	SessionID  string                 `json:"session_id,omitempty"`
	DeviceID   string                 `json:"device_id,omitempty"`
	IP         string                 `json:"ip,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	Timestamp  time.Time              `json:"timestamp"`
}
