package model

import "time"

type DeletionRequest struct {
	UserID           int64  `json:"user_id"`
	VerificationCode string `json:"verification_code" validate:"required"`
	VerificationType string `json:"verification_type" validate:"required,oneof=sms_code email_otp"`
}

type Deletion struct {
	UserID      int64      `json:"user_id" db:"user_id"`
	RequestedAt *time.Time `json:"requested_at,omitempty" db:"requested_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty" db:"cancelled_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

type DeletionResponse struct {
	Message string `json:"message"`
}

const FreezePeriod = 7 * 24 * time.Hour
