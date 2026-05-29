package repository

import (
	"context"
	"database/sql"
)

type RoleRepository interface {
	GetUserRoles(ctx context.Context, accountID string) ([]string, error)
	GetUserRolesByUserID(ctx context.Context, userID int64) (string, []string, error)
}

type roleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) GetUserRoles(ctx context.Context, accountID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.name FROM roles r
		JOIN user_roles ur ON ur.role_id = r.id
		WHERE ur.user_id = $1
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		roles = append(roles, name)
	}
	if roles == nil {
		roles = []string{}
	}
	return roles, rows.Err()
}

func (r *roleRepository) GetUserRolesByUserID(ctx context.Context, userID int64) (string, []string, error) {
	var accountID string
	err := r.db.QueryRowContext(ctx, `SELECT account_id FROM public.users WHERE id = $1`, userID).Scan(&accountID)
	if err != nil {
		return "", nil, err
	}
	roles, err := r.GetUserRoles(ctx, accountID)
	if err != nil {
		return "", nil, err
	}
	return accountID, roles, nil
}
