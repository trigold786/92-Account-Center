package model

import "time"

// Refund represents a refund request and its lifecycle state.
type Refund struct {
	ID               int64      `json:"id"`
	OrderID          int64      `json:"order_id"`
	UserID           int64      `json:"user_id"`
	Amount           float64    `json:"amount"`
	Reason           string     `json:"reason"`
	Status           string     `json:"status"`
	RefundNo         string     `json:"refund_no,omitempty"`
	Provider         string     `json:"provider,omitempty"`
	ProviderRefundID string     `json:"provider_refund_id,omitempty"`
	ProviderStatus   string     `json:"provider_status,omitempty"`
	ProviderError    string     `json:"provider_error,omitempty"`
	ApproverID       int64      `json:"approver_id,omitempty"`
	ReviewNote       string     `json:"review_note,omitempty"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty"`
	FailedAt         *time.Time `json:"failed_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}
