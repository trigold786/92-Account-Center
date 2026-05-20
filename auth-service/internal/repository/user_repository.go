package repository

import (
	"context"
	"database/sql"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type UserRepository interface {
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByAccountID(ctx context.Context, accountID string) (*model.User, error)
	UpdatePasswordHash(ctx context.Context, userID int64, passwordHash string) error
	FindBySocialAccount(ctx context.Context, provider, providerUID string) (*model.User, error)
	CreateFromSocial(ctx context.Context, info *model.SocialUserInfo) (*model.User, error)
	LinkSocialAccount(ctx context.Context, userID int64, provider, providerUID, email, avatar, accessToken, refreshToken string) error
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, phone_number, account_id, email, password_hash, mfa_enabled, mfa_secret, created_at, updated_at FROM users WHERE phone_number = $1`
	err := r.db.QueryRowContext(ctx, query, phoneNumber).Scan(
		&user.ID, &user.PhoneNumber, &user.AccountID, &user.Email,
		&user.PasswordHash, &user.MFAEnabled, &user.MFASecret,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, phone_number, account_id, email, password_hash, mfa_enabled, mfa_secret, created_at, updated_at FROM users WHERE email = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.PhoneNumber, &user.AccountID, &user.Email,
		&user.PasswordHash, &user.MFAEnabled, &user.MFASecret,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) GetByAccountID(ctx context.Context, accountID string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, phone_number, account_id, email, password_hash, mfa_enabled, mfa_secret, created_at, updated_at FROM users WHERE account_id = $1`
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&user.ID, &user.PhoneNumber, &user.AccountID, &user.Email,
		&user.PasswordHash, &user.MFAEnabled, &user.MFASecret,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) UpdatePasswordHash(ctx context.Context, userID int64, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID, passwordHash)
	return err
}

func (r *userRepository) FindBySocialAccount(ctx context.Context, provider, providerUID string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT u.id, u.phone_number, u.account_id, u.email, u.password_hash,
	                  u.mfa_enabled, u.mfa_secret, u.last_strong_auth_at, u.created_at, u.updated_at
	           FROM users u
	           INNER JOIN social_accounts sa ON sa.user_id = u.id
	           WHERE sa.provider = $1 AND sa.provider_uid = $2`
	err := r.db.QueryRowContext(ctx, query, provider, providerUID).Scan(
		&user.ID, &user.PhoneNumber, &user.AccountID, &user.Email,
		&user.PasswordHash, &user.MFAEnabled, &user.MFASecret,
		&user.LastStrongAuthAt, &user.CreatedAt, &user.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) CreateFromSocial(ctx context.Context, info *model.SocialUserInfo) (*model.User, error) {
	user := &model.User{}
	err := r.db.QueryRowContext(ctx,
		`INSERT INTO users (email, created_at, updated_at) VALUES ($1, NOW(), NOW()) RETURNING id, account_id`,
		info.Email,
	).Scan(&user.ID, &user.AccountID)
	if err != nil {
		return nil, err
	}
	user.Email = &info.Email
	return user, nil
}

func (r *userRepository) LinkSocialAccount(ctx context.Context, userID int64, provider, providerUID, email, avatar, accessToken, refreshToken string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO social_accounts (user_id, provider, provider_uid, email, avatar_url, access_token, refresh_token, created_at, updated_at)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),NOW())
		 ON CONFLICT (provider, provider_uid) DO UPDATE SET
		   email = EXCLUDED.email, avatar_url = EXCLUDED.avatar_url,
		   access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token,
		   updated_at = NOW()`,
		userID, provider, providerUID, email, avatar, accessToken, refreshToken)
	return err
}
