package model

import "time"

type Refund struct {
	ID         int64     `json:"id"`
	OrderID    int64     `json:"order_id"`
	UserID     int64     `json:"user_id"`
	Amount     float64   `json:"amount"`
	Reason     string    `json:"reason"`
	Status     string    `json:"status"`
	ApproverID int64     `json:"approver_id,omitempty"`
	ReviewNote string    `json:"review_note,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
