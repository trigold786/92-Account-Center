package model

import "time"

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRefunded  OrderStatus = "refunded"
)

type Order struct {
	ID                   int64       `json:"id" db:"id"`
	OrderNo              string      `json:"order_no" db:"order_no"`
	UserID               int64       `json:"user_id" db:"user_id"`
	ProductType          string      `json:"product_type" db:"product_type"`
	ProductName          string      `json:"product_name" db:"product_name"`
	Amount               float64     `json:"amount" db:"amount"`
	Currency             string      `json:"currency" db:"currency"`
	Status               OrderStatus `json:"status" db:"status"`
	PaymentMethod        string      `json:"payment_method,omitempty" db:"payment_method"`
	PaymentTransactionID string      `json:"payment_transaction_id,omitempty" db:"payment_transaction_id"`
	PaidAt               *time.Time  `json:"paid_at,omitempty" db:"paid_at"`
	CancelledAt          *time.Time  `json:"cancelled_at,omitempty" db:"cancelled_at"`
	RefundedAt           *time.Time  `json:"refunded_at,omitempty" db:"refunded_at"`
	RefundReason         string      `json:"refund_reason,omitempty" db:"refund_reason"`
	ExpiresAt            *time.Time  `json:"expires_at,omitempty" db:"expires_at"`
	Metadata             string      `json:"metadata,omitempty" db:"metadata"`
	CreatedAt            time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time   `json:"updated_at" db:"updated_at"`
}

type CreateOrderRequest struct {
	UserID      int64      `json:"user_id" binding:"required"`
	ProductType string     `json:"product_type" binding:"required"`
	ProductName string     `json:"product_name" binding:"required"`
	Amount      float64    `json:"amount" binding:"required,gt=0"`
	Currency    string     `json:"currency"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Metadata    string     `json:"metadata,omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status              OrderStatus `json:"status" binding:"required"`
	PaymentMethod       string      `json:"payment_method,omitempty"`
	PaymentTransactionID string     `json:"payment_transaction_id,omitempty"`
	RefundReason        string      `json:"refund_reason,omitempty"`
}

type OrderQueryRequest struct {
	UserID       *int64       `form:"user_id"`
	Status       *OrderStatus `form:"status"`
	PaymentMethod string      `form:"payment_method"`
	StartTime    *time.Time   `form:"start_time"`
	EndTime      *time.Time   `form:"end_time"`
	Page         int          `form:"page"`
	PageSize     int          `form:"page_size"`
}

type OrderListResponse struct {
	Orders   []Order `json:"orders"`
	Total    int     `json:"total"`
	Page     int     `json:"page"`
	PageSize int     `json:"page_size"`
}
