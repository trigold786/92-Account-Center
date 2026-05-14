package model

import "time"

type BlacklistEntry struct {
	ID         int64      `json:"id" db:"id"`
	EntryType  string     `json:"entry_type" db:"entry_type"`
	EntryValue string     `json:"entry_value" db:"entry_value"`
	Reason     string     `json:"reason" db:"reason"`
	CreatedBy  string     `json:"created_by" db:"created_by"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty" db:"expires_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

type BlacklistEntryRequest struct {
	EntryType  string `json:"entry_type" binding:"required,oneof=IP DEVICE PHONE ACCOUNT"`
	EntryValue string `json:"entry_value" binding:"required"`
	Reason     string `json:"reason" binding:"required"`
	CreatedBy  string `json:"created_by"`
	ExpiresAt  string `json:"expires_at,omitempty"`
}

type BlacklistCheckRequest struct {
	EntryType  string `json:"entry_type" binding:"required,oneof=IP DEVICE PHONE ACCOUNT"`
	EntryValue string `json:"entry_value" binding:"required"`
}
