package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

type ConfigRepository interface {
	// Groups
	ListGroups(ctx context.Context) ([]model.ConfigGroup, error)
	GetGroupByID(ctx context.Context, id int64) (*model.ConfigGroup, error)
	CreateGroup(ctx context.Context, g *model.ConfigGroup) error
	UpdateGroup(ctx context.Context, g *model.ConfigGroup) error
	DeleteGroup(ctx context.Context, id int64) error

	// Items
	ListItems(ctx context.Context, filter model.ConfigItemFilter) ([]model.ConfigItem, int, error)
	GetItemByID(ctx context.Context, id int64) (*model.ConfigItem, error)
	GetItemByCode(ctx context.Context, code string) (*model.ConfigItem, error)
	CreateItem(ctx context.Context, item *model.ConfigItem) error
	UpdateItem(ctx context.Context, item *model.ConfigItem) error
	DeleteItem(ctx context.Context, id int64) error
	ResetItemToDefault(ctx context.Context, id int64) error

	// Stats
	GetTotalCount(ctx context.Context) (int, error)

	// Versions
	ListVersionsByItemID(ctx context.Context, itemID int64) ([]model.ConfigVersion, error)
	CreateVersion(ctx context.Context, v *model.ConfigVersion) error
}

type configRepository struct {
	db *sql.DB
}

func NewConfigRepository(db *sql.DB) ConfigRepository {
	return &configRepository{db: db}
}

// --- Groups ---

func (r *configRepository) ListGroups(ctx context.Context) ([]model.ConfigGroup, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM config_groups ORDER BY id`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []model.ConfigGroup
	for rows.Next() {
		var g model.ConfigGroup
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	return groups, rows.Err()
}

func (r *configRepository) GetGroupByID(ctx context.Context, id int64) (*model.ConfigGroup, error) {
	g := &model.ConfigGroup{}
	query := `SELECT id, name, description, created_at, updated_at FROM config_groups WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&g.ID, &g.Name, &g.Description, &g.CreatedAt, &g.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return g, err
}

func (r *configRepository) CreateGroup(ctx context.Context, g *model.ConfigGroup) error {
	query := `INSERT INTO config_groups (name, description) VALUES ($1, $2) RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query, g.Name, g.Description).Scan(&g.ID, &g.CreatedAt, &g.UpdatedAt)
}

func (r *configRepository) UpdateGroup(ctx context.Context, g *model.ConfigGroup) error {
	query := `UPDATE config_groups SET name = $2, description = $3, updated_at = NOW() WHERE id = $1 RETURNING updated_at`
	return r.db.QueryRowContext(ctx, query, g.ID, g.Name, g.Description).Scan(&g.UpdatedAt)
}

func (r *configRepository) DeleteGroup(ctx context.Context, id int64) error {
	query := `DELETE FROM config_groups WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// --- Items ---

func (r *configRepository) ListItems(ctx context.Context, filter model.ConfigItemFilter) ([]model.ConfigItem, int, error) {
	where := []string{}
	args := []interface{}{}
	argIdx := 1

	if filter.GroupID != nil {
		where = append(where, fmt.Sprintf("group_id = $%d", argIdx))
		args = append(args, *filter.GroupID)
		argIdx++
	}
	if filter.Code != "" {
		where = append(where, fmt.Sprintf("code ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Code+"%")
		argIdx++
	}
	if filter.Name != "" {
		where = append(where, fmt.Sprintf("name ILIKE $%d", argIdx))
		args = append(args, "%"+filter.Name+"%")
		argIdx++
	}
	if filter.DataType != "" {
		where = append(where, fmt.Sprintf("data_type = $%d", argIdx))
		args = append(args, filter.DataType)
		argIdx++
	}
	if filter.IsEnabled != nil {
		where = append(where, fmt.Sprintf("is_enabled = $%d", argIdx))
		args = append(args, *filter.IsEnabled)
		argIdx++
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = " WHERE " + strings.Join(where, " AND ")
	}

	countQuery := "SELECT COUNT(*) FROM config_items" + whereClause
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	dataQuery := fmt.Sprintf(`SELECT id, group_id, code, name, description, data_type, current_value,
		default_value, COALESCE(min_value,''), COALESCE(max_value,''), COALESCE(allowed_values,''), is_sensitive, is_enabled, created_at, updated_at
		FROM config_items%s ORDER BY group_id, code LIMIT $%d OFFSET $%d`, whereClause, argIdx, argIdx+1)
	args = append(args, filter.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []model.ConfigItem
	for rows.Next() {
		var item model.ConfigItem
		if err := rows.Scan(&item.ID, &item.GroupID, &item.Code, &item.Name, &item.Description,
			&item.DataType, &item.CurrentValue, &item.DefaultValue, &item.MinValue, &item.MaxValue,
			&item.AllowedValues, &item.IsSensitive, &item.IsEnabled, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *configRepository) GetItemByID(ctx context.Context, id int64) (*model.ConfigItem, error) {
	item := &model.ConfigItem{}
	query := `SELECT id, group_id, code, name, description, data_type, current_value,
		default_value, COALESCE(min_value,''), COALESCE(max_value,''), COALESCE(allowed_values,''), is_sensitive, is_enabled, created_at, updated_at
		FROM config_items WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&item.ID, &item.GroupID, &item.Code, &item.Name,
		&item.Description, &item.DataType, &item.CurrentValue, &item.DefaultValue, &item.MinValue,
		&item.MaxValue, &item.AllowedValues, &item.IsSensitive, &item.IsEnabled, &item.CreatedAt, &item.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

func (r *configRepository) GetItemByCode(ctx context.Context, code string) (*model.ConfigItem, error) {
	item := &model.ConfigItem{}
	query := `SELECT id, group_id, code, name, description, data_type, current_value,
		default_value, COALESCE(min_value,''), COALESCE(max_value,''), COALESCE(allowed_values,''), is_sensitive, is_enabled, created_at, updated_at
		FROM config_items WHERE code = $1`
	err := r.db.QueryRowContext(ctx, query, code).Scan(&item.ID, &item.GroupID, &item.Code, &item.Name,
		&item.Description, &item.DataType, &item.CurrentValue, &item.DefaultValue, &item.MinValue,
		&item.MaxValue, &item.AllowedValues, &item.IsSensitive, &item.IsEnabled, &item.CreatedAt, &item.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return item, err
}

func (r *configRepository) CreateItem(ctx context.Context, item *model.ConfigItem) error {
	query := `INSERT INTO config_items (group_id, code, name, description, data_type, current_value,
		default_value, min_value, max_value, allowed_values, is_sensitive, is_enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12) RETURNING id, created_at, updated_at`
	return r.db.QueryRowContext(ctx, query,
		item.GroupID, item.Code, item.Name, item.Description, item.DataType,
		item.CurrentValue, item.DefaultValue, item.MinValue, item.MaxValue,
		item.AllowedValues, item.IsSensitive, item.IsEnabled,
	).Scan(&item.ID, &item.CreatedAt, &item.UpdatedAt)
}

func (r *configRepository) UpdateItem(ctx context.Context, item *model.ConfigItem) error {
	query := `UPDATE config_items SET group_id = $2, name = $3, description = $4, data_type = $5,
		current_value = $6, default_value = $7, min_value = $8, max_value = $9, allowed_values = $10,
		is_sensitive = $11, is_enabled = $12, updated_at = NOW() WHERE id = $1 RETURNING updated_at`
	return r.db.QueryRowContext(ctx, query,
		item.ID, item.GroupID, item.Name, item.Description, item.DataType,
		item.CurrentValue, item.DefaultValue, item.MinValue, item.MaxValue,
		item.AllowedValues, item.IsSensitive, item.IsEnabled,
	).Scan(&item.UpdatedAt)
}

func (r *configRepository) DeleteItem(ctx context.Context, id int64) error {
	query := `DELETE FROM config_items WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *configRepository) ResetItemToDefault(ctx context.Context, id int64) error {
	query := `UPDATE config_items SET current_value = default_value, updated_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// --- Versions ---

func (r *configRepository) ListVersionsByItemID(ctx context.Context, itemID int64) ([]model.ConfigVersion, error) {
	query := `SELECT id, item_id, value_before, value_after, change_reason, changed_by, created_at
		FROM config_versions WHERE item_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []model.ConfigVersion
	for rows.Next() {
		var v model.ConfigVersion
		if err := rows.Scan(&v.ID, &v.ItemID, &v.ValueBefore, &v.ValueAfter, &v.ChangeReason, &v.ChangedBy, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

func (r *configRepository) GetTotalCount(ctx context.Context) (int, error) {
	var total int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM config_items").Scan(&total)
	return total, err
}

func (r *configRepository) CreateVersion(ctx context.Context, v *model.ConfigVersion) error {
	query := `INSERT INTO config_versions (item_id, value_before, value_after, change_reason, changed_by)
		VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRowContext(ctx, query, v.ItemID, v.ValueBefore, v.ValueAfter, v.ChangeReason, v.ChangedBy).Scan(&v.ID, &v.CreatedAt)
}
