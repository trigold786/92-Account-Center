package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
)

type BlacklistRepository struct {
	db *sql.DB
}

func NewBlacklistRepository(db *sql.DB) *BlacklistRepository {
	return &BlacklistRepository{db: db}
}

func (r *BlacklistRepository) Create(ctx context.Context, entry *model.BlacklistEntry) error {
	query := `INSERT INTO blacklist_entries (entry_type, entry_value, reason, created_by, expires_at)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query,
		entry.EntryType, entry.EntryValue, entry.Reason, entry.CreatedBy, entry.ExpiresAt,
	).Scan(&entry.ID, &entry.CreatedAt)
}

func (r *BlacklistRepository) CheckBlocked(ctx context.Context, entryType, entryValue string) (*model.BlacklistEntry, error) {
	entry := &model.BlacklistEntry{}
	query := `SELECT id, entry_type, entry_value, reason, created_by, expires_at, created_at
		FROM blacklist_entries
		WHERE entry_type = $1 AND entry_value = $2
		AND (expires_at IS NULL OR expires_at > $3)
		LIMIT 1`
	err := r.db.QueryRowContext(ctx, query, entryType, entryValue, time.Now()).Scan(
		&entry.ID, &entry.EntryType, &entry.EntryValue, &entry.Reason,
		&entry.CreatedBy, &entry.ExpiresAt, &entry.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func (r *BlacklistRepository) Remove(ctx context.Context, entryType, entryValue string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM blacklist_entries WHERE entry_type = $1 AND entry_value = $2`, entryType, entryValue)
	return err
}

func (r *BlacklistRepository) List(ctx context.Context, entryType string, limit, offset int) ([]model.BlacklistEntry, error) {
	query := `SELECT id, entry_type, entry_value, reason, created_by, expires_at, created_at
		FROM blacklist_entries WHERE ($1 = '' OR entry_type = $1)
		ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, entryType, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var entries []model.BlacklistEntry
	for rows.Next() {
		var e model.BlacklistEntry
		if err := rows.Scan(&e.ID, &e.EntryType, &e.EntryValue, &e.Reason, &e.CreatedBy, &e.ExpiresAt, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}
