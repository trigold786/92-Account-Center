package model

import (
	"time"
)

type User struct {
	ID                  int64      `json:"id" db:"id"`
	PhoneNumber         string     `json:"phone_number" db:"phone_number"`
	AccountID           string     `json:"account_id" db:"account_id"`
	Email               string     `json:"email,omitempty" db:"email"`
	PasswordHash        string     `json:"-" db:"password_hash"`
	MFAEnabled          bool       `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret           string     `json:"-" db:"mfa_secret"`
	LastStrongAuthAt    *time.Time `json:"last_strong_auth_at,omitempty" db:"last_strong_auth_at"`
	IdentityTier        int        `json:"identity_tier" db:"identity_tier"`
	Status              string     `json:"status" db:"status"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	DeletionRequestedAt *time.Time `json:"deletion_requested_at,omitempty" db:"deletion_requested_at"`
	DeletionExpiresAt   *time.Time `json:"deletion_expires_at,omitempty" db:"deletion_expires_at"`
	DeletionCancelledAt *time.Time `json:"deletion_cancelled_at,omitempty" db:"deletion_cancelled_at"`
	DeletionDeletedAt   *time.Time `json:"deletion_deleted_at,omitempty" db:"deletion_deleted_at"`
}

type UserProfile struct {
	ID          int64     `json:"id"`
	PhoneNumber string    `json:"phone_number"`
	AccountID   string    `json:"account_id"`
	Email       string    `json:"email"`
	MFAEnabled  bool      `json:"mfa_enabled"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}
