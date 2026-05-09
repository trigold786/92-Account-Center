package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           int64     `json:"id" db:"id"`
	PhoneNumber  string    `json:"phone_number" db:"phone_number"`
	AccountID    string    `json:"account_id" db:"account_id"`
	PasswordHash string    `json:"password_hash" db:"password_hash"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// TableName returns the table name for User
func (User) TableName() string {
	return "users"
}