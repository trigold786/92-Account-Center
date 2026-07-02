package model

import "time"

type PaymentCallback struct {
	ID            int64     `json:"id" db:"id"`
	Provider      string    `json:"provider" db:"provider"`
	OrderNo       string    `json:"order_no" db:"order_no"`
	TransactionID string    `json:"transaction_id" db:"transaction_id"`
	Status        string    `json:"status" db:"status"`
	Verified      bool      `json:"verified" db:"verified"`
	RawPayload    string    `json:"raw_payload" db:"raw_payload"`
	ReceivedAt    time.Time `json:"received_at" db:"received_at"`
}
