package model

import (
	"time"

	"github.com/google/uuid"
)

type SubAccountRole string

const (
	SubAccountRoleAdmin  SubAccountRole = "admin"
	SubAccountRoleMember SubAccountRole = "member"
)

type SubAccountStatus string

const (
	SubAccountStatusActive    SubAccountStatus = "active"
	SubAccountStatusSuspended SubAccountStatus = "suspended"
	SubAccountStatusPending   SubAccountStatus = "pending_liveness"
)

type SubAccount struct {
	ID             uuid.UUID        `json:"id" db:"id"`
	EnterpriseID   uuid.UUID        `json:"enterprise_id" db:"enterprise_id"`
	UserID         uuid.UUID        `json:"user_id" db:"user_id"`
	Role           SubAccountRole   `json:"role" db:"role"`
	Status         SubAccountStatus `json:"status" db:"status"`
	LastLivenessAt *time.Time       `json:"last_liveness_at,omitempty" db:"last_liveness_at"`
	CreatedAt      time.Time        `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at" db:"updated_at"`
}

type CreateSubAccountRequest struct {
	EnterpriseID string `json:"enterprise_id" binding:"required"`
	UserID       string `json:"user_id" binding:"required"`
	Role         string `json:"role" binding:"required,oneof=admin member"`
}

type UpdateSubAccountRequest struct {
	Role   string `json:"role,omitempty"`
	Status string `json:"status,omitempty"`
}

type SubAccountResponse struct {
	ID             string `json:"id"`
	EnterpriseID   string `json:"enterprise_id"`
	UserID         string `json:"user_id"`
	Role           string `json:"role"`
	Status         string `json:"status"`
	LastLivenessAt string `json:"last_liveness_at,omitempty"`
}

type RequireLivenessRequest struct {
	SubAccountID string `json:"sub_account_id" binding:"required"`
	Reason       string `json:"reason"`
}

type CompleteLivenessRequest struct {
	SubAccountID string `json:"sub_account_id" binding:"required"`
	Token        string `json:"token" binding:"required"`
}
