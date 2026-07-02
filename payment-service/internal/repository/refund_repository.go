package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
)

type RefundRepository struct {
	db *sql.DB
}

func NewRefundRepository(db *sql.DB) *RefundRepository {
	return &RefundRepository{db: db}
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanRefund(row rowScanner) (*model.Refund, error) {
	ref := &model.Refund{}
	err := row.Scan(
		&ref.ID, &ref.OrderID, &ref.UserID, &ref.Amount, &ref.Reason, &ref.Status,
		&ref.RefundNo, &ref.Provider, &ref.ProviderRefundID, &ref.ProviderStatus, &ref.ProviderError,
		&ref.ApproverID, &ref.ReviewNote, &ref.ApprovedAt, &ref.FailedAt,
		&ref.CreatedAt, &ref.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func (r *RefundRepository) Create(ctx context.Context, refund *model.Refund) error {
	refund.CreatedAt = time.Now()
	refund.UpdatedAt = refund.CreatedAt
	return r.db.QueryRowContext(ctx,
		`INSERT INTO refunds (order_id, user_id, amount, reason, status, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		refund.OrderID, refund.UserID, refund.Amount, refund.Reason, refund.Status, refund.CreatedAt, refund.UpdatedAt,
	).Scan(&refund.ID)
}

func (r *RefundRepository) GetByID(ctx context.Context, id int64) (*model.Refund, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, order_id, user_id, amount, reason, status, refund_no, provider, provider_refund_id, provider_status, provider_error, COALESCE(approver_id,0), review_note, approved_at, failed_at, created_at, updated_at
		 FROM refunds WHERE id=$1`, id)
	ref, err := scanRefund(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return ref, nil
}

func (r *RefundRepository) UpdateStatus(ctx context.Context, id int64, status string, approverID int64, note string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refunds SET status=$1, approver_id=$2, review_note=$3, approved_at=CASE WHEN $1='approved' THEN NOW() ELSE approved_at END, updated_at=NOW() WHERE id=$4`,
		status, approverID, note, id)
	return err
}

func (r *RefundRepository) FindByOrderID(ctx context.Context, orderID int64) (*model.Refund, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, order_id, user_id, amount, reason, status, refund_no, provider, provider_refund_id, provider_status, provider_error, COALESCE(approver_id,0), review_note, approved_at, failed_at, created_at, updated_at
		 FROM refunds WHERE order_id=$1 LIMIT 1`, orderID)
	ref, err := scanRefund(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return ref, nil
}

func (r *RefundRepository) ListByUserID(ctx context.Context, userID int64) ([]*model.Refund, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, order_id, user_id, amount, reason, status, refund_no, provider, provider_refund_id, provider_status, provider_error, COALESCE(approver_id,0), review_note, approved_at, failed_at, created_at, updated_at
		 FROM refunds WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refunds []*model.Refund
	for rows.Next() {
		ref, err := scanRefund(rows)
		if err != nil {
			return nil, err
		}
		refunds = append(refunds, ref)
	}
	return refunds, nil
}

func (r *RefundRepository) UpdateProviderResult(ctx context.Context, id int64, refundNo string, providerName string, providerRefundID string, providerStatus string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refunds SET refund_no=$1, provider=$2, provider_refund_id=$3, provider_status=$4, provider_error='', updated_at=NOW() WHERE id=$5`, refundNo, providerName, providerRefundID, providerStatus, id)
	return err
}

func (r *RefundRepository) MarkProviderFailure(ctx context.Context, id int64, refundNo string, providerName string, providerError string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE refunds SET status='failed', refund_no=$1, provider=$2, provider_error=$3, failed_at=NOW(), updated_at=NOW() WHERE id=$4`, refundNo, providerName, providerError, id)
	return err
}
