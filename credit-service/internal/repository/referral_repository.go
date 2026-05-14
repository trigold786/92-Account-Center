package repository

import (
	"context"
	"database/sql"

	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
)

type ReferralRepository interface {
	Create(ctx context.Context, referrerID, refereeID int64) (*model.ReferralRelation, error)
	GetByRefereeID(ctx context.Context, refereeID int64) (*model.ReferralRelation, error)
	GetByReferrerID(ctx context.Context, referrerID int64) ([]model.ReferralRelation, error)
	IncrementSubscriptionCount(ctx context.Context, refereeID int64) error
	GetReferralSummary(ctx context.Context, referrerID int64) (*model.ReferralSummary, error)
}

type referralRepository struct {
	db *sql.DB
}

func NewReferralRepository(db *sql.DB) ReferralRepository {
	return &referralRepository{db: db}
}

func (r *referralRepository) Create(ctx context.Context, referrerID, refereeID int64) (*model.ReferralRelation, error) {
	rel := &model.ReferralRelation{}
	query := `
		INSERT INTO referral_relations (referrer_id, referee_id, referee_subscription_count, status, created_at, updated_at)
		VALUES ($1, $2, 0, 'ACTIVE', NOW(), NOW())
		RETURNING id, referrer_id, referee_id, referee_subscription_count, status, created_at, updated_at`
	err := r.db.QueryRowContext(ctx, query, referrerID, refereeID).Scan(
		&rel.ID, &rel.ReferrerID, &rel.RefereeID, &rel.RefereeSubscriptionCount,
		&rel.Status, &rel.CreatedAt, &rel.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return rel, nil
}

func (r *referralRepository) GetByRefereeID(ctx context.Context, refereeID int64) (*model.ReferralRelation, error) {
	rel := &model.ReferralRelation{}
	query := `SELECT id, referrer_id, referee_id, referee_subscription_count, status, created_at, updated_at
		FROM referral_relations WHERE referee_id = $1`
	err := r.db.QueryRowContext(ctx, query, refereeID).Scan(
		&rel.ID, &rel.ReferrerID, &rel.RefereeID, &rel.RefereeSubscriptionCount,
		&rel.Status, &rel.CreatedAt, &rel.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return rel, nil
}

func (r *referralRepository) GetByReferrerID(ctx context.Context, referrerID int64) ([]model.ReferralRelation, error) {
	query := `SELECT id, referrer_id, referee_id, referee_subscription_count, status, created_at, updated_at
		FROM referral_relations WHERE referrer_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, referrerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var relations []model.ReferralRelation
	for rows.Next() {
		var rel model.ReferralRelation
		if err := rows.Scan(
			&rel.ID, &rel.ReferrerID, &rel.RefereeID, &rel.RefereeSubscriptionCount,
			&rel.Status, &rel.CreatedAt, &rel.UpdatedAt,
		); err != nil {
			return nil, err
		}
		relations = append(relations, rel)
	}
	return relations, rows.Err()
}

func (r *referralRepository) IncrementSubscriptionCount(ctx context.Context, refereeID int64) error {
	query := `UPDATE referral_relations SET referee_subscription_count = referee_subscription_count + 1, updated_at = NOW() WHERE referee_id = $1`
	_, err := r.db.ExecContext(ctx, query, refereeID)
	return err
}

func (r *referralRepository) GetReferralSummary(ctx context.Context, referrerID int64) (*model.ReferralSummary, error) {
	summary := &model.ReferralSummary{}
	query := `SELECT
		COUNT(*) as total_referees,
		COALESCE(SUM(CASE WHEN ct.type = 'REFERRAL_BONUS' THEN ct.amount ELSE 0 END), 0) as total_earned,
		COUNT(CASE WHEN rr.status = 'ACTIVE' THEN 1 END) as active_referees
		FROM referral_relations rr
		LEFT JOIN credit_accounts ca ON rr.referee_id = ca.user_id
		LEFT JOIN credit_transactions ct ON ca.id = ct.credit_account_id AND ct.type = 'REFERRAL_BONUS'
		WHERE rr.referrer_id = $1`
	err := r.db.QueryRowContext(ctx, query, referrerID).Scan(
		&summary.TotalReferees, &summary.TotalEarned, &summary.ActiveReferees,
	)
	if err != nil {
		return nil, err
	}
	return summary, nil
}
