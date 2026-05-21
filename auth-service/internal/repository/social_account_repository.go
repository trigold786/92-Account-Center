package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type SocialAccountRepository struct {
	db *sql.DB
}

func NewSocialAccountRepository(db *sql.DB) *SocialAccountRepository {
	return &SocialAccountRepository{db: db}
}

func (r *SocialAccountRepository) FindByProvider(ctx context.Context, provider, providerUID string) (*model.SocialAccount, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, provider, provider_uid, COALESCE(email,''), COALESCE(avatar_url,''),
		        COALESCE(access_token,''), COALESCE(refresh_token,''), created_at, updated_at
		 FROM social_accounts WHERE provider=$1 AND provider_uid=$2`, provider, providerUID)
	sa := &model.SocialAccount{}
	err := row.Scan(&sa.ID, &sa.UserID, &sa.Provider, &sa.ProviderUID, &sa.Email, &sa.AvatarURL,
		&sa.AccessToken, &sa.RefreshToken, &sa.CreatedAt, &sa.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return sa, nil
}

func (r *SocialAccountRepository) FindByUserID(ctx context.Context, userID int64) ([]*model.SocialAccount, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, user_id, provider, provider_uid, COALESCE(email,''), COALESCE(avatar_url,''),
		        created_at, updated_at
		 FROM social_accounts WHERE user_id=$1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var accounts []*model.SocialAccount
	for rows.Next() {
		sa := &model.SocialAccount{}
		if err := rows.Scan(&sa.ID, &sa.UserID, &sa.Provider, &sa.ProviderUID, &sa.Email, &sa.AvatarURL,
			&sa.CreatedAt, &sa.UpdatedAt); err != nil {
			return nil, err
		}
		accounts = append(accounts, sa)
	}
	return accounts, nil
}

func (r *SocialAccountRepository) Create(ctx context.Context, sa *model.SocialAccount) error {
	sa.CreatedAt = time.Now()
	sa.UpdatedAt = sa.CreatedAt
	return r.db.QueryRowContext(ctx,
		`INSERT INTO social_accounts (user_id, provider, provider_uid, email, avatar_url, access_token, refresh_token, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		sa.UserID, sa.Provider, sa.ProviderUID, sa.Email, sa.AvatarURL, sa.AccessToken, sa.RefreshToken, sa.CreatedAt, sa.UpdatedAt,
	).Scan(&sa.ID)
}

func (r *SocialAccountRepository) Delete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM social_accounts WHERE id=$1`, id)
	return err
}
