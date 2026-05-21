package model

import "time"

type GuestSession struct {
	ID        int64     `json:"id"`
	AccountID string    `json:"account_id"`
	DeviceID  string    `json:"device_id"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateGuestRequest struct {
	DeviceID string `json:"device_id"`
}

type UpgradeGuestRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
}
