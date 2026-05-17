package model

import "time"

type AuditLog struct {
	ID              int64     `json:"id"`
	OperationType   string    `json:"operation_type"`
	OperationObject string    `json:"operation_object"`
	Operator        string    `json:"operator"`
	OperatorIP      string    `json:"operator_ip"`
	OperationResult string    `json:"operation_result"`
	OperationDetails string   `json:"operation_details"`
	SM3Hash         string    `json:"sm3_hash"`
	CreatedAt       time.Time `json:"created_at"`
}

type AuditLogFilter struct {
	OperationType string
	Operator      string
	StartTime     *time.Time
	EndTime       *time.Time
	Page          int
	PageSize      int
}
