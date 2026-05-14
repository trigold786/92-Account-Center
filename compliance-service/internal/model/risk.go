package model

import (
	"encoding/json"
	"time"
)

type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

type EventType string

const (
	EventTypeLogin              EventType = "login"
	EventTypeImpossibleTravel   EventType = "impossible_travel"
	EventTypeDeviceChange       EventType = "device_change"
	EventTypeVelocityExceeded   EventType = "velocity_exceeded"
	EventTypeRiskScoreExceeded  EventType = "risk_score_exceeded"
)

type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type RiskEvent struct {
	RiskEventID string          `json:"risk_event_id" db:"risk_event_id"`
	UserID      string          `json:"user_id" db:"user_id"`
	EventType   EventType       `json:"event_type" db:"event_type"`
	RiskScore   int             `json:"risk_score" db:"risk_score"`
	RiskLevel   RiskLevel       `json:"risk_level" db:"risk_level"`
	Details     json.RawMessage `json:"details" db:"details"`
	IPAddress   string          `json:"ip_address" db:"ip_address"`
	Location    *Location       `json:"location" db:"location"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
}

type LoginContext struct {
	UserID            string    `json:"user_id"`
	IP                string    `json:"ip"`
	DeviceFingerprint string    `json:"device_fingerprint"`
	Timestamp         time.Time `json:"timestamp"`
	Location          *Location `json:"location"`
}
