package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID                  int64     `json:"id" db:"id"`
	PhoneNumber         string    `json:"phone_number" db:"phone_number"`
	AccountID           string    `json:"account_id" db:"account_id"`
	PasswordHash        string    `json:"password_hash" db:"password_hash"`
	CreatedAt           time.Time `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time `json:"updated_at" db:"updated_at"`
	DeletionRequestedAt time.Time `json:"deletion_requested_at" db:"deletion_requested_at"`
	DeletionExpiresAt   time.Time `json:"deletion_expires_at" db:"deletion_expires_at"`
	DeletionCancelledAt time.Time `json:"deletion_cancelled_at" db:"deletion_cancelled_at"`
	DeletionDeletedAt   time.Time `json:"deletion_deleted_at" db:"deletion_deleted_at"`
}

// TableName returns the table name for User
func (User) TableName() string {
	return "users"
}