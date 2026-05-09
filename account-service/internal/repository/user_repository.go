package repository

import (
	"context"
	"database/sql"

	"github.com/sunxi/92-Account-Center/account-service/internal/model"
)

// UserRepository defines the interface for user data access.
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error)
	GetByAccountID(ctx context.Context, accountID string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, userID string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	ExistsByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error)
	ExistsByAccountID(ctx context.Context, accountID string) (bool, error)
	PermanentDelete(ctx context.Context, userID int64) error
}

// userRepository implements UserRepository using PostgreSQL.
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// Create inserts a new user into the database.
func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	query := `
		INSERT INTO users (phone_number, account_id, password_hash, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		user.PhoneNumber,
		user.AccountID,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
}

// GetByPhoneNumber retrieves a user by phone number.
func (r *userRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, phone_number, account_id, password_hash, created_at, updated_at FROM users WHERE phone_number = $1`
	err := r.db.QueryRowContext(ctx, query, phoneNumber).Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.AccountID,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// GetByAccountID retrieves a user by account ID.
func (r *userRepository) GetByAccountID(ctx context.Context, accountID string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, phone_number, account_id, password_hash, created_at, updated_at FROM users WHERE account_id = $1`
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.AccountID,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// GetByEmail retrieves a user by email (phone number in our schema).
// Note: In our current schema, we use phone_number for email as well for simplicity.
// A production system would have a separate email field.
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, phone_number, account_id, password_hash, created_at, updated_at FROM users WHERE phone_number = $1`
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.AccountID,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

// ExistsByPhoneNumber checks if a user with the given phone number exists.
func (r *userRepository) ExistsByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone_number = $1)`
	err := r.db.QueryRowContext(ctx, query, phoneNumber).Scan(&exists)
	return exists, err
}

// ExistsByAccountID checks if a user with the given account ID exists.
func (r *userRepository) ExistsByAccountID(ctx context.Context, accountID string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE account_id = $1)`
	err := r.db.QueryRowContext(ctx, query, accountID).Scan(&exists)
	return exists, err
}

// GetByID retrieves a user by ID.
func (r *userRepository) GetByID(ctx context.Context, userID string) (*model.User, error) {
	var id int64
	query := `SELECT id, phone_number, account_id, password_hash, created_at, updated_at FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&id,
		&user.PhoneNumber,
		&user.AccountID,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	user.ID = id
	return user, nil
}

// Update updates an existing user in the database.
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	query := `
		UPDATE users 
		SET phone_number = $2, account_id = $3, password_hash = $4, 
			updated_at = $5, deletion_requested_at = $6, 
			deletion_expires_at = $7, deletion_cancelled_at = $8, 
			deletion_deleted_at = $9
		WHERE id = $1
		RETURNING updated_at`
	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.PhoneNumber,
		user.AccountID,
		user.PasswordHash,
		time.Now(),
		user.DeletionRequestedAt,
		user.DeletionExpiresAt,
		user.DeletionCancelledAt,
		user.DeletionDeletedAt,
	).Scan(&user.UpdatedAt)
	return err
}

// PermanentDelete permanently removes a user from the database.
func (r *userRepository) PermanentDelete(ctx context.Context, userID int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}