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
		`SELECT id, order_id, user_id, amount, reason, status, COALESCE(approver_id,0), COALESCE(review_note,''), created_at, updated_at
		 FROM refunds WHERE id=$1`, id)
	ref := &model.Refund{}
	err := row.Scan(&ref.ID, &ref.OrderID, &ref.UserID, &ref.Amount, &ref.Reason, &ref.Status, &ref.ApproverID, &ref.ReviewNote, &ref.CreatedAt, &ref.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return ref, nil
}

func (r *RefundRepository) UpdateStatus(ctx context.Context, id int64, status string, approverID int64, note string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE refunds SET status=$1, approver_id=$2, review_note=$3, updated_at=NOW() WHERE id=$4`,
		status, approverID, note, id)
	return err
}

func (r *RefundRepository) FindByOrderID(ctx context.Context, orderID int64) (*model.Refund, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, order_id, user_id, amount, reason, status, COALESCE(approver_id,0), COALESCE(review_note,''), created_at, updated_at
		 FROM refunds WHERE order_id=$1 LIMIT 1`, orderID)
	ref := &model.Refund{}
	err := row.Scan(&ref.ID, &ref.OrderID, &ref.UserID, &ref.Amount, &ref.Reason, &ref.Status, &ref.ApproverID, &ref.ReviewNote, &ref.CreatedAt, &ref.UpdatedAt)
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
		`SELECT id, order_id, user_id, amount, reason, status, COALESCE(approver_id,0), created_at, updated_at
		 FROM refunds WHERE user_id=$1 ORDER BY created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var refunds []*model.Refund
	for rows.Next() {
		ref := &model.Refund{}
		if err := rows.Scan(&ref.ID, &ref.OrderID, &ref.UserID, &ref.Amount, &ref.Reason, &ref.Status, &ref.ApproverID, &ref.CreatedAt, &ref.UpdatedAt); err != nil {
			return nil, err
		}
		refunds = append(refunds, ref)
	}
	return refunds, nil
}
