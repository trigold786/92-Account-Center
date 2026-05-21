package model

import "time"

type Promotion struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	DiscountPct float64   `json:"discount_pct"`
	PlanIDs     []int64   `json:"plan_ids"`
	StartAt     time.Time `json:"start_at"`
	EndAt       time.Time `json:"end_at"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
}
