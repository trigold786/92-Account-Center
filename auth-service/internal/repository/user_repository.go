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
