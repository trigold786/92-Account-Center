package model

import "time"

type Coupon struct {
	ID            int64      `json:"id"`
	Code          string     `json:"code"`
	DiscountType  string     `json:"discount_type"`
	DiscountValue float64    `json:"discount_value"`
	MaxUses       int        `json:"max_uses"`
	CurrentUses   int        `json:"current_uses"`
	MaxPerUser    int        `json:"max_per_user"`
	Active        bool       `json:"active"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}
