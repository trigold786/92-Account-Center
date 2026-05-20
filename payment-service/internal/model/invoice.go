package model

import "time"

type Invoice struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	OrderID   int64     `json:"order_id"`
	InvoiceNo string   `json:"invoice_no"`
	Title     string   `json:"title"`
	TaxID     string   `json:"tax_id,omitempty"`
	Email     string   `json:"email"`
	Amount    float64  `json:"amount"`
	Status    string   `json:"status"`
	FileURL   string   `json:"file_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
