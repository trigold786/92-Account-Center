package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

type InvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) Create(ctx context.Context, inv *model.Invoice) error {
	inv.CreatedAt = time.Now()
	inv.UpdatedAt = inv.CreatedAt
	return r.db.QueryRowContext(ctx,
		`INSERT INTO invoices (user_id, order_id, invoice_no, title, tax_id, email, amount, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		inv.UserID, inv.OrderID, inv.InvoiceNo, inv.Title, inv.TaxID, inv.Email, inv.Amount, inv.Status, inv.CreatedAt, inv.UpdatedAt,
	).Scan(&inv.ID)
}

func (r *InvoiceRepository) GetByUserID(ctx context.Context, userID int64) ([]*model.Invoice, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, order_id, invoice_no, title, COALESCE(tax_id,''), email, amount, status, COALESCE(file_url,''), created_at, updated_at
		 FROM invoices WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var invoices []*model.Invoice
	for rows.Next() {
		inv := &model.Invoice{}
		if err := rows.Scan(&inv.ID, &inv.UserID, &inv.OrderID, &inv.InvoiceNo, &inv.Title, &inv.TaxID, &inv.Email,
			&inv.Amount, &inv.Status, &inv.FileURL, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, nil
}
