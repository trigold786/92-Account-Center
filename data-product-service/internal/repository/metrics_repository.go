package repository

import (
	"context"
	"database/sql"
	"time"
)

type DateCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type MetricsRepository struct {
	db *sql.DB
}

func NewMetricsRepository(db *sql.DB) *MetricsRepository {
	return &MetricsRepository{db: db}
}

func (r *MetricsRepository) GetRegistrationCountByPeriod(ctx context.Context, period string) ([]DateCount, error) {
	var interval string
	switch period {
	case "daily":
		interval = "1 day"
	case "weekly":
		interval = "1 week"
	case "monthly":
		interval = "1 month"
	default:
		interval = "1 day"
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT date_trunc($1, created_at)::date AS date, COUNT(*) AS count
		 FROM users GROUP BY date ORDER BY date DESC LIMIT 30`, interval)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var counts []DateCount
	for rows.Next() {
		var dc DateCount
		var t time.Time
		if err := rows.Scan(&t, &dc.Count); err != nil {
			return nil, err
		}
		dc.Date = t.Format("2006-01-02")
		counts = append(counts, dc)
	}
	return counts, nil
}

func (r *MetricsRepository) GetPaidUsers(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users WHERE subscription_status='paid'`).Scan(&count)
	return count, err
}

func (r *MetricsRepository) GetMRR(ctx context.Context) (float64, error) {
	var total float64
	err := r.db.QueryRowContext(ctx, `SELECT COALESCE(SUM(amount),0) FROM orders WHERE status='paid'`).Scan(&total)
	return total, err
}

func (r *MetricsRepository) GetTotalUsers(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (r *MetricsRepository) GetSubscribedUsers(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions WHERE status='active'`).Scan(&count)
	return count, err
}
