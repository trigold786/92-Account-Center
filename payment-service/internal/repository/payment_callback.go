package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

type PaymentCallbackRepository interface {
	Save(ctx context.Context, callback *model.PaymentCallback) error
}

type paymentCallbackRepository struct {
	db *sql.DB
}

func NewPaymentCallbackRepository(db *sql.DB) PaymentCallbackRepository {
	return &paymentCallbackRepository{db: db}
}

func (r *paymentCallbackRepository) Save(ctx context.Context, callback *model.PaymentCallback) error {
	if callback.ReceivedAt.IsZero() {
		callback.ReceivedAt = time.Now()
	}
	query := `
		INSERT INTO payment_callbacks (provider, order_no, transaction_id, status, verified, raw_payload, received_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`
	return r.db.QueryRowContext(ctx, query,
		callback.Provider,
		callback.OrderNo,
		callback.TransactionID,
		callback.Status,
		callback.Verified,
		callback.RawPayload,
		callback.ReceivedAt,
	).Scan(&callback.ID)
}
