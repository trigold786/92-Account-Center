package repository

import (
	"context"
	"database/sql"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

type EntitlementRepository interface {
	Create(ctx context.Context, e *model.Entitlement) error
	GetByUserID(ctx context.Context, userID int64) ([]model.Entitlement, error)
	GetByUserAndFeature(ctx context.Context, userID int64, featureCode string) (*model.Entitlement, error)
	UpdateQuota(ctx context.Context, id int64, usedQuota int) error
	UpdateTotalQuota(ctx context.Context, id int64, totalQuota int) error
	DeleteByUserID(ctx context.Context, userID int64) error
}

type entitlementRepository struct {
	db *sql.DB
}

func NewEntitlementRepository(db *sql.DB) EntitlementRepository {
	return &entitlementRepository{db: db}
}

func (r *entitlementRepository) Create(ctx context.Context, e *model.Entitlement) error {
	query := `
		INSERT INTO entitlements (user_id, feature_code, total_quota, used_quota, reset_time, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		e.UserID, e.FeatureCode, e.TotalQuota, e.UsedQuota, e.ResetTime,
	).Scan(&e.ID, &e.CreatedAt, &e.UpdatedAt)
}

func (r *entitlementRepository) GetByUserID(ctx context.Context, userID int64) ([]model.Entitlement, error) {
	query := `SELECT id, user_id, feature_code, total_quota, used_quota, reset_time, created_at, updated_at
		FROM entitlements WHERE user_id = $1`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entitlements []model.Entitlement
	for rows.Next() {
		var e model.Entitlement
		if err := rows.Scan(&e.ID, &e.UserID, &e.FeatureCode, &e.TotalQuota, &e.UsedQuota, &e.ResetTime, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entitlements = append(entitlements, e)
	}
	return entitlements, rows.Err()
}

func (r *entitlementRepository) GetByUserAndFeature(ctx context.Context, userID int64, featureCode string) (*model.Entitlement, error) {
	query := `SELECT id, user_id, feature_code, total_quota, used_quota, reset_time, created_at, updated_at
		FROM entitlements WHERE user_id = $1 AND feature_code = $2`
	var e model.Entitlement
	err := r.db.QueryRowContext(ctx, query, userID, featureCode).Scan(
		&e.ID, &e.UserID, &e.FeatureCode, &e.TotalQuota, &e.UsedQuota, &e.ResetTime, &e.CreatedAt, &e.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *entitlementRepository) UpdateQuota(ctx context.Context, id int64, usedQuota int) error {
	query := `UPDATE entitlements SET used_quota = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, usedQuota)
	return err
}

func (r *entitlementRepository) UpdateTotalQuota(ctx context.Context, id int64, totalQuota int) error {
	query := `UPDATE entitlements SET total_quota = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id, totalQuota)
	return err
}

func (r *entitlementRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	query := `DELETE FROM entitlements WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}
