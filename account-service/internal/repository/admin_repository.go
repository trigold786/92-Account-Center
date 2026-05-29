package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

type AdminRepository interface {
	ListUsers(ctx context.Context, page, pageSize int, search string, status string, tier int) ([]model.User, int, error)
	GetUserDetail(ctx context.Context, userID int64) (*model.User, error)
	UpdateUserStatus(ctx context.Context, userID int64, status string, reason string, operator string) error
	AdjustIdentityTier(ctx context.Context, userID int64, tier int, reason string, operator string) error
	InsertAuditLog(ctx context.Context, userID int64, action string, details string, operator string) error
	GetAuditLog(ctx context.Context, userID int64, page, pageSize int) ([]model.AuditLogEntry, int, error)
	EnsureAuditLogTable(ctx context.Context) error
}

type adminRepository struct {
	db *sql.DB
}

func NewAdminRepository(db *sql.DB) AdminRepository {
	return &adminRepository{db: db}
}

func (r *adminRepository) EnsureAuditLogTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS audit_log (
			id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL,
			action VARCHAR(255) NOT NULL,
			details TEXT,
			operator VARCHAR(255),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)`
	_, err := r.db.ExecContext(ctx, query)
	return err
}

func (r *adminRepository) ListUsers(ctx context.Context, page, pageSize int, search string, status string, tier int) ([]model.User, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if search != "" {
		conditions = append(conditions, fmt.Sprintf("(phone_number LIKE $%d OR account_id LIKE $%d OR email LIKE $%d)", argIdx, argIdx, argIdx))
		args = append(args, "%"+search+"%")
		argIdx++
	}
	if status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, status)
		argIdx++
	}
	if tier >= 0 {
		conditions = append(conditions, fmt.Sprintf("identity_tier = $%d", argIdx))
		args = append(args, tier)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = " WHERE " + strings.Join(conditions, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM users" + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		slog.Error("admin count query failed", "query", countQuery, "args", args, "error", err.Error())
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	listQuery := `SELECT id, phone_number, account_id, email, password_hash, mfa_enabled, mfa_secret,
		last_strong_auth_at, identity_tier, status, created_at, updated_at,
		deletion_requested_at, deletion_expires_at, deletion_cancelled_at, deletion_deleted_at
		FROM users` + whereClause + fmt.Sprintf(" ORDER BY id DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, listQuery, args...)
	if err != nil {
		slog.Error("admin list query failed", "query", listQuery, "args", args, "error", err.Error())
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(
			&u.ID, &u.PhoneNumber, &u.AccountID, &u.Email, &u.PasswordHash,
			&u.MFAEnabled, &u.MFASecret, &u.LastStrongAuthAt,
			&u.IdentityTier, &u.Status, &u.CreatedAt, &u.UpdatedAt,
			&u.DeletionRequestedAt, &u.DeletionExpiresAt,
			&u.DeletionCancelledAt, &u.DeletionDeletedAt,
		); err != nil {
			slog.Error("admin scan user failed", "error", err.Error())
			return nil, 0, err
		}
		users = append(users, u)
	}
	return users, total, rows.Err()
}

func (r *adminRepository) GetUserDetail(ctx context.Context, userID int64) (*model.User, error) {
	user := &model.User{}
	query := `SELECT id, phone_number, account_id, email, password_hash, mfa_enabled, mfa_secret,
		last_strong_auth_at, identity_tier, status, created_at, updated_at,
		deletion_requested_at, deletion_expires_at, deletion_cancelled_at, deletion_deleted_at
		FROM users WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.PhoneNumber, &user.AccountID, &user.Email, &user.PasswordHash,
		&user.MFAEnabled, &user.MFASecret, &user.LastStrongAuthAt,
		&user.IdentityTier, &user.Status, &user.CreatedAt, &user.UpdatedAt,
		&user.DeletionRequestedAt, &user.DeletionExpiresAt,
		&user.DeletionCancelledAt, &user.DeletionDeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}

func (r *adminRepository) UpdateUserStatus(ctx context.Context, userID int64, status string, reason string, operator string) error {
	query := `UPDATE users SET status = $2, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, userID, status)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	details, _ := json.Marshal(map[string]string{"status": status, "reason": reason})
	return r.InsertAuditLog(ctx, userID, "update_status", string(details), operator)
}

func (r *adminRepository) AdjustIdentityTier(ctx context.Context, userID int64, tier int, reason string, operator string) error {
	query := `UPDATE users SET identity_tier = $2, updated_at = NOW() WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, userID, tier)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}

	details, _ := json.Marshal(map[string]interface{}{"tier": tier, "reason": reason})
	return r.InsertAuditLog(ctx, userID, "adjust_tier", string(details), operator)
}

func (r *adminRepository) InsertAuditLog(ctx context.Context, userID int64, action string, details string, operator string) error {
	query := `INSERT INTO audit_log (user_id, action, details, operator) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, userID, action, details, operator)
	return err
}

func (r *adminRepository) GetAuditLog(ctx context.Context, userID int64, page, pageSize int) ([]model.AuditLogEntry, int, error) {
	countQuery := `SELECT COUNT(*) FROM audit_log WHERE user_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := `SELECT id, user_id, action, details, operator, created_at
		FROM audit_log WHERE user_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var entries []model.AuditLogEntry
	for rows.Next() {
		var e model.AuditLogEntry
		if err := rows.Scan(&e.ID, &e.UserID, &e.Action, &e.Details, &e.Operator, &e.CreatedAt); err != nil {
			return nil, 0, err
		}
		entries = append(entries, e)
	}
	return entries, total, rows.Err()
}
