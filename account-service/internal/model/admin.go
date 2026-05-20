package model

import "time"

type AdminUserListRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Search   string `form:"search"`
	Status   string `form:"status"`
	Tier     int    `form:"tier"`
}

type AdminUserListResponse struct {
	Users    []User `json:"users"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type AdminStatusUpdateRequest struct {
	Status string `json:"status" binding:"required,oneof=active frozen banned"`
	Reason string `json:"reason" binding:"required"`
}

type AdminTierUpdateRequest struct {
	Tier   int    `json:"tier" binding:"required,oneof=0 1 2 3 4"`
	Reason string `json:"reason" binding:"required"`
}

type AdminCreditAdjustRequest struct {
	Amount int64  `json:"amount" binding:"required"`
	Reason string `json:"reason" binding:"required"`
	Type   string `json:"type" binding:"required,oneof=earn consume"`
}

type AuditLogEntry struct {
	ID        int64     `json:"id" db:"id"`
	UserID    int64     `json:"user_id" db:"user_id"`
	Action    string    `json:"action" db:"action"`
	Details   string    `json:"details" db:"details"`
	Operator  string    `json:"operator" db:"operator"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
