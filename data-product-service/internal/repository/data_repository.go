package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
)

type DataRepository interface {
	GetSubscriptionStats(ctx context.Context, userID int64) (*model.SubscriptionStats, error)
	GetAllSubscriptionStats(ctx context.Context) (map[int64]*model.SubscriptionStats, error)
	GetTotalUsers(ctx context.Context) (int, error)
	GetTotalSubscriptions(ctx context.Context) (int, error)
	GetTotalCreditsByTypes(ctx context.Context, types []string) (float64, error)
	GetActiveBlacklistCount(ctx context.Context) (int, error)
	GetRegistrationTrend(ctx context.Context, days int) ([]model.DailyCount, error)
	GetCreditFlow(ctx context.Context) (map[string]float64, error)
	GetUserTierCounts(ctx context.Context) ([]model.UserTierCount, error)
	GetDistinctSubscriberCount(ctx context.Context, minTier int) (int, error)
}

type dataRepository struct {
	db *sql.DB
}

func NewDataRepository(db *sql.DB) DataRepository {
	return &dataRepository{db: db}
}

func (r *dataRepository) GetSubscriptionStats(ctx context.Context, userID int64) (*model.SubscriptionStats, error) {
	stats := &model.SubscriptionStats{}
	query := `SELECT COUNT(*), COALESCE(SUM(price), 0), COALESCE(TO_CHAR(MAX(end_time), 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '') FROM subscriptions WHERE user_id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&stats.Freq, &stats.Monetary, &stats.LastSubAt)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *dataRepository) GetAllSubscriptionStats(ctx context.Context) (map[int64]*model.SubscriptionStats, error) {
	query := `SELECT user_id, COUNT(*), COALESCE(SUM(price), 0), COALESCE(TO_CHAR(MAX(end_time), 'YYYY-MM-DD"T"HH24:MI:SS"Z"'), '') FROM subscriptions GROUP BY user_id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64]*model.SubscriptionStats)
	for rows.Next() {
		var uid int64
		stats := &model.SubscriptionStats{}
		if err := rows.Scan(&uid, &stats.Freq, &stats.Monetary, &stats.LastSubAt); err != nil {
			return nil, err
		}
		result[uid] = stats
	}
	return result, rows.Err()
}

func (r *dataRepository) GetTotalUsers(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}

func (r *dataRepository) GetTotalSubscriptions(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM subscriptions`).Scan(&count)
	return count, err
}

func (r *dataRepository) GetTotalCreditsByTypes(ctx context.Context, types []string) (float64, error) {
	var total float64
	query := `SELECT COALESCE(SUM(amount), 0) FROM credit_transactions WHERE type = ANY($1) AND status IN ('AVAILABLE', 'CONSUMED')`
	err := r.db.QueryRowContext(ctx, query, pq.Array(types)).Scan(&total)
	return total, err
}

func (r *dataRepository) GetActiveBlacklistCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM blacklist_entries`).Scan(&count)
	return count, err
}

func (r *dataRepository) GetRegistrationTrend(ctx context.Context, days int) ([]model.DailyCount, error) {
	query := fmt.Sprintf(`SELECT TO_CHAR(created_at, 'YYYY-MM-DD') AS d, COUNT(*) FROM users WHERE created_at >= NOW() - interval '%d days' GROUP BY d ORDER BY d DESC`, days)
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.DailyCount
	for rows.Next() {
		var dc model.DailyCount
		if err := rows.Scan(&dc.Date, &dc.Count); err != nil {
			return nil, err
		}
		result = append(result, dc)
	}
	return result, rows.Err()
}

func (r *dataRepository) GetCreditFlow(ctx context.Context) (map[string]float64, error) {
	query := `SELECT type, COALESCE(SUM(amount), 0) FROM credit_transactions WHERE status IN ('AVAILABLE', 'CONSUMED') GROUP BY type`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]float64)
	for rows.Next() {
		var t string
		var amt float64
		if err := rows.Scan(&t, &amt); err != nil {
			return nil, err
		}
		result[t] = amt
	}
	return result, rows.Err()
}

func (r *dataRepository) GetUserTierCounts(ctx context.Context) ([]model.UserTierCount, error) {
	query := `SELECT identity_tier, COUNT(*) FROM users GROUP BY identity_tier ORDER BY identity_tier`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []model.UserTierCount
	for rows.Next() {
		var tc model.UserTierCount
		if err := rows.Scan(&tc.Tier, &tc.Count); err != nil {
			return nil, err
		}
		result = append(result, tc)
	}
	return result, rows.Err()
}

func (r *dataRepository) GetDistinctSubscriberCount(ctx context.Context, minTier int) (int, error) {
	var count int
	query := `SELECT COUNT(DISTINCT user_id) FROM subscriptions WHERE status = 'ACTIVE'`
	if minTier > 0 {
		query += ` AND tier_level >= $1`
		err := r.db.QueryRowContext(ctx, query, minTier).Scan(&count)
		return count, err
	}
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}
