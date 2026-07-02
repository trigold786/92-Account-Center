package repository

import (
	"context"
	"database/sql"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

type SubscriptionRepository interface {
	Create(ctx context.Context, sub *model.Subscription) error
	GetActiveByUserID(ctx context.Context, userID int64) (*model.Subscription, error)
	GetByUserID(ctx context.Context, userID int64) ([]model.Subscription, error)
	GetByOrderID(ctx context.Context, orderID string) (*model.Subscription, error)
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdateEndTime(ctx context.Context, id int64, endTime string) error
	FindExpired(ctx context.Context) ([]model.Subscription, error)
}

type subscriptionRepository struct {
	db *sql.DB
}

func NewSubscriptionRepository(db *sql.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	query := `
		INSERT INTO subscriptions (user_id, tier_level, start_time, end_time, status, price, payment_method, order_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		sub.UserID, sub.TierLevel, sub.StartTime, sub.EndTime, sub.Status, sub.Price, sub.PaymentMethod, sub.OrderID,
	).Scan(&sub.ID, &sub.CreatedAt, &sub.UpdatedAt)
}

func (r *subscriptionRepository) GetActiveByUserID(ctx context.Context, userID int64) (*model.Subscription, error) {
	query := `SELECT id, user_id, tier_level, start_time, end_time, status, price, payment_method, order_id, created_at, updated_at
		FROM subscriptions WHERE user_id = $1 AND status = 'ACTIVE' ORDER BY created_at DESC LIMIT 1`
	var sub model.Subscription
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&sub.ID, &sub.UserID, &sub.TierLevel, &sub.StartTime, &sub.EndTime, &sub.Status, &sub.Price, &sub.PaymentMethod, &sub.OrderID, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) GetByUserID(ctx context.Context, userID int64) ([]model.Subscription, error) {
	query := `SELECT id, user_id, tier_level, start_time, end_time, status, price, payment_method, order_id, created_at, updated_at
		FROM subscriptions WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.TierLevel, &sub.StartTime, &sub.EndTime, &sub.Status, &sub.Price, &sub.PaymentMethod, &sub.OrderID, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

func (r *subscriptionRepository) GetByOrderID(ctx context.Context, orderID string) (*model.Subscription, error) {
	query := `SELECT id, user_id, tier_level, start_time, end_time, status, price, payment_method, order_id, created_at, updated_at
		FROM subscriptions WHERE order_id = $1 ORDER BY created_at DESC LIMIT 1`
	var sub model.Subscription
	err := r.db.QueryRowContext(ctx, query, orderID).Scan(
		&sub.ID, &sub.UserID, &sub.TierLevel, &sub.StartTime, &sub.EndTime, &sub.Status, &sub.Price, &sub.PaymentMethod, &sub.OrderID, &sub.CreatedAt, &sub.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	query := `UPDATE subscriptions SET status = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, status)
	return err
}

func (r *subscriptionRepository) UpdateEndTime(ctx context.Context, id int64, endTime string) error {
	query := `UPDATE subscriptions SET end_time = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, endTime)
	return err
}

func (r *subscriptionRepository) FindExpired(ctx context.Context) ([]model.Subscription, error) {
	query := `SELECT id, user_id, tier_level, start_time, end_time, status, price, payment_method, order_id, created_at, updated_at
		FROM subscriptions WHERE status = 'ACTIVE' AND end_time < NOW()`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []model.Subscription
	for rows.Next() {
		var sub model.Subscription
		if err := rows.Scan(&sub.ID, &sub.UserID, &sub.TierLevel, &sub.StartTime, &sub.EndTime, &sub.Status, &sub.Price, &sub.PaymentMethod, &sub.OrderID, &sub.CreatedAt, &sub.UpdatedAt); err != nil {
			return nil, err
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}
