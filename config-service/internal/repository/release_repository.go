package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

type ReleaseRepository interface {
	ListReleases(ctx context.Context, status string, page, pageSize int) ([]model.ConfigRelease, int, error)
	GetReleaseByID(ctx context.Context, id int64) (*model.ConfigRelease, error)
	CreateRelease(ctx context.Context, r *model.ConfigRelease) error
	UpdateReleaseStatus(ctx context.Context, id int64, status, approvedBy string) error

	ListReleaseItems(ctx context.Context, releaseID int64) ([]model.ConfigReleaseItem, error)
	AddReleaseItem(ctx context.Context, ri *model.ConfigReleaseItem) error
	RemoveReleaseItem(ctx context.Context, id int64) error
}

type releaseRepository struct {
	db *sql.DB
}

func NewReleaseRepository(db *sql.DB) ReleaseRepository {
	return &releaseRepository{db: db}
}

func (r *releaseRepository) ListReleases(ctx context.Context, status string, page, pageSize int) ([]model.ConfigRelease, int, error) {
	where := ""
	args := []interface{}{}
	if status != "" {
		where = " WHERE status = $1"
		args = append(args, status)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM config_releases"+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`SELECT id, title, description, status, created_by, approved_by,
		created_at, updated_at, approved_at, released_at
		FROM config_releases%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		where, len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var releases []model.ConfigRelease
	for rows.Next() {
		var rel model.ConfigRelease
		if err := rows.Scan(&rel.ID, &rel.Title, &rel.Description, &rel.Status, &rel.CreatedBy,
			&rel.ApprovedBy, &rel.CreatedAt, &rel.UpdatedAt, &rel.ApprovedAt, &rel.ReleasedAt); err != nil {
			return nil, 0, err
		}
		releases = append(releases, rel)
	}
	return releases, total, rows.Err()
}

func (r *releaseRepository) GetReleaseByID(ctx context.Context, id int64) (*model.ConfigRelease, error) {
	rel := &model.ConfigRelease{}
	query := `SELECT id, title, description, status, created_by, approved_by,
		created_at, updated_at, approved_at, released_at FROM config_releases WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&rel.ID, &rel.Title, &rel.Description, &rel.Status,
		&rel.CreatedBy, &rel.ApprovedBy, &rel.CreatedAt, &rel.UpdatedAt, &rel.ApprovedAt, &rel.ReleasedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return rel, err
}

func (r *releaseRepository) CreateRelease(ctx context.Context, rel *model.ConfigRelease) error {
	query := `INSERT INTO config_releases (title, description, status, created_by)
		VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, rel.Title, rel.Description, rel.Status, rel.CreatedBy).
		Scan(&rel.ID, &rel.CreatedAt, &rel.UpdatedAt)
}

func (r *releaseRepository) UpdateReleaseStatus(ctx context.Context, id int64, status, approvedBy string) error {
	var query string
	if status == "approved" {
		query = `UPDATE config_releases SET status = $2, approved_by = $3, approved_at = NOW(), updated_at = NOW() WHERE id = $1`
	} else if status == "released" {
		query = `UPDATE config_releases SET status = $2, released_at = NOW(), updated_at = NOW() WHERE id = $1`
	} else {
		query = `UPDATE config_releases SET status = $2, updated_at = NOW() WHERE id = $1`
	}
	_, err := r.db.ExecContext(ctx, query, id, status, approvedBy)
	return err
}

func (r *releaseRepository) ListReleaseItems(ctx context.Context, releaseID int64) ([]model.ConfigReleaseItem, error) {
	query := `SELECT id, release_id, item_id, value_before, value_after, change_reason, created_at
		FROM config_release_items WHERE release_id = $1`
	rows, err := r.db.QueryContext(ctx, query, releaseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.ConfigReleaseItem
	for rows.Next() {
		var ri model.ConfigReleaseItem
		if err := rows.Scan(&ri.ID, &ri.ReleaseID, &ri.ItemID, &ri.ValueBefore, &ri.ValueAfter, &ri.ChangeReason, &ri.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, ri)
	}
	return items, rows.Err()
}

func (r *releaseRepository) AddReleaseItem(ctx context.Context, ri *model.ConfigReleaseItem) error {
	query := `INSERT INTO config_release_items (release_id, item_id, value_before, value_after, change_reason)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, ri.ReleaseID, ri.ItemID, ri.ValueBefore, ri.ValueAfter, ri.ChangeReason).
		Scan(&ri.ID, &ri.CreatedAt)
}

func (r *releaseRepository) RemoveReleaseItem(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM config_release_items WHERE id = $1", id)
	return err
}
