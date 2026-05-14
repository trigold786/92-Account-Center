package model

import (
	"encoding/json"
	"time"
)

type AuditLogEntry struct {
	UserID         *int64          `json:"user_id"`
	ActionType     string          `json:"action_type" binding:"required"`
	TargetResource string          `json:"target_resource" binding:"required"`
	SourceIP       string          `json:"source_ip" binding:"required"`
	Result         string          `json:"result" binding:"required,oneof=success failure"`
	Details        json.RawMessage `json:"details"`
	EventTime      *time.Time      `json:"event_time"`
}

type BatchAuditLogRequest struct {
	Entries []AuditLogEntry `json:"entries" binding:"required,min=1,max=1000"`
}

type AuditLogResponse struct {
	LogID   string `json:"log_id"`
	SM3Hash string `json:"sm3_hash"`
}

type BatchAuditLogResponse struct {
	Processed int                `json:"processed"`
	Failed    int                `json:"failed"`
	Logs      []AuditLogResponse `json:"logs"`
}

type LogQueryParams struct {
	UserID    string     `form:"user_id"`
	StartTime *time.Time `form:"start_time" time_format:"2006-01-02T15:04:05Z07:00"`
	EndTime   *time.Time `form:"end_time" time_format:"2006-01-02T15:04:05Z07:00"`
	Limit     int        `form:"limit,default=100"`
	Offset    int        `form:"offset,default=0"`
}

type IntegrityVerifyResponse struct {
	LogID         string `json:"log_id"`
	IsValid       bool   `json:"is_valid"`
	StoredHash    string `json:"stored_hash"`
	ComputedHash  string `json:"computed_hash"`
}

type CleanupResponse struct {
	DeletedCount  int64 `json:"deleted_count"`
	RetentionDays int   `json:"retention_days"`
}
