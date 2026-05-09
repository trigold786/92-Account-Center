package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	LogID           string          `json:"log_id" db:"log_id"`
	UserID          *int64          `json:"user_id" db:"user_id"`
	EventTime       time.Time       `json:"event_time" db:"event_time"`
	ActionType      string          `json:"action_type" db:"action_type"`
	TargetResource  string          `json:"target_resource" db:"target_resource"`
	SourceIP        string          `json:"source_ip" db:"source_ip"`
	Result          string          `json:"result" db:"result"`
	Details         json.RawMessage `json:"details" db:"details"`
	SM3Hash         string          `json:"sm3_hash" db:"sm3_hash"`
	CreatedAt       time.Time       `json:"created_at" db:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

func NewAuditLog(userID *int64, actionType, targetResource, sourceIP, result string, details json.RawMessage) *AuditLog {
	return &AuditLog{
		LogID:          uuid.New().String(),
		UserID:         userID,
		EventTime:      time.Now(),
		ActionType:     actionType,
		TargetResource: targetResource,
		SourceIP:       sourceIP,
		Result:         result,
		Details:        details,
	}
}
