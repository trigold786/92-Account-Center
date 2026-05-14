package model

import "time"

type Subscription struct {
	ID            int64     `json:"id"`
	UserID        int64     `json:"user_id"`
	TierLevel     int       `json:"tier_level"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	Status        string    `json:"status"`
	Price         float64   `json:"price"`
	PaymentMethod string    `json:"payment_method,omitempty"`
	OrderID       string    `json:"order_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PurchaseRequest struct {
	UserID        string  `json:"user_id" binding:"required"`
	TierLevel     int     `json:"tier_level" binding:"required,oneof=2 3 4"`
	Price         float64 `json:"price" binding:"required"`
	PaymentMethod string  `json:"payment_method"`
}

type UpgradeRequest struct {
	UserID        string  `json:"user_id" binding:"required"`
	NewTier       int     `json:"new_tier" binding:"required,oneof=3 4"`
	PriceDiff     float64 `json:"price_diff" binding:"required"`
	PaymentMethod string  `json:"payment_method"`
}

type RenewRequest struct {
	UserID        string `json:"user_id" binding:"required"`
	PaymentMethod string `json:"payment_method"`
}
