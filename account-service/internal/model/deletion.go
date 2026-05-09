package model

import (
	"time"
)

// DeletionRequest represents a request to delete an account
type DeletionRequest struct {
	UserID           int64  `json:"user_id"`
	VerificationCode string `json:"verification_code" validate:"required"`
	VerificationType string `json:"verification_type" validate:"required,oneof=sms_code email_otp"`
}

// Deletion represents an account deletion record
type Deletion struct {
	UserID         int64     `json:"user_id" db:"user_id"`
	RequestedAt    time.Time `json:"requested_at" db:"requested_at"`
	ExpiresAt      time.Time `json:"expires_at" db:"expires_at"`
	CancelledAt    time.Time `json:"cancelled_at" db:"cancelled_at"`
	DeletedAt      time.Time `json:"deleted_at" db:"deleted_at"`
}

// DeletionResponse represents the response for deletion operations
type DeletionResponse struct {
	Message string `json:"message"`
}

// FreezePeriod is the default account freeze period before permanent deletion
const FreezePeriod = 7 * 24 * time.Hour // 7 days