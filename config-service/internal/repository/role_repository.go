package repository

import (
	"context"
	"database/sql"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

type RoleRepository interface {
	ListRoles(ctx context.Context) ([]model.Role, error)
	GetRoleByID(ctx context.Context, id int64) (*model.Role, error)
	CreateRole(ctx context.Context, r *model.Role) error
	DeleteRole(ctx context.Context, id int64) error

	GetRolePermissions(ctx context.Context, roleID int64) ([]model.RolePermission, error)
	AddRolePermission(ctx context.Context, rp *model.RolePermission) error
	RemoveRolePermission(ctx context.Context, id int64) error

	GetUserRoles(ctx context.Context, userID string) ([]model.UserRole, error)
	SetUserRole(ctx context.Context, ur *model.UserRole) error
	RemoveUserRole(ctx context.Context, userID string, roleID int64) error
}

type roleRepository struct {
	db *sql.DB
}

func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepository{db: db}
}

func (r *roleRepository) ListRoles(ctx context.Context) ([]model.Role, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name, description, created_at FROM roles ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []model.Role
	for rows.Next() {
		var role model.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (r *roleRepository) GetRoleByID(ctx context.Context, id int64) (*model.Role, error) {
	role := &model.Role{}
	err := r.db.QueryRowContext(ctx, "SELECT id, name, description, created_at FROM roles WHERE id = $1", id).
		Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return role, err
}

func (r *roleRepository) CreateRole(ctx context.Context, role *model.Role) error {
	return r.db.QueryRowContext(ctx, "INSERT INTO roles (name, description) VALUES ($1, $2) RETURNING id, created_at",
		role.Name, role.Description).Scan(&role.ID, &role.CreatedAt)
}

func (r *roleRepository) GetRolePermissions(ctx context.Context, roleID int64) ([]model.RolePermission, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, role_id, permission, created_at FROM role_permissions WHERE role_id = $1", roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perms []model.RolePermission
	for rows.Next() {
		var rp model.RolePermission
		if err := rows.Scan(&rp.ID, &rp.RoleID, &rp.Permission, &rp.CreatedAt); err != nil {
			return nil, err
		}
		perms = append(perms, rp)
	}
	return perms, rows.Err()
}

func (r *roleRepository) AddRolePermission(ctx context.Context, rp *model.RolePermission) error {
	return r.db.QueryRowContext(ctx, "INSERT INTO role_permissions (role_id, permission) VALUES ($1, $2) RETURNING id, created_at",
		rp.RoleID, rp.Permission).Scan(&rp.ID, &rp.CreatedAt)
}

func (r *roleRepository) RemoveRolePermission(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM role_permissions WHERE id = $1", id)
	return err
}

func (r *roleRepository) GetUserRoles(ctx context.Context, userID string) ([]model.UserRole, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, user_id, role_id, created_at FROM user_roles WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urs []model.UserRole
	for rows.Next() {
		var ur model.UserRole
		if err := rows.Scan(&ur.ID, &ur.UserID, &ur.RoleID, &ur.CreatedAt); err != nil {
			return nil, err
		}
		urs = append(urs, ur)
	}
	return urs, rows.Err()
}

func (r *roleRepository) SetUserRole(ctx context.Context, ur *model.UserRole) error {
	return r.db.QueryRowContext(ctx, "INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2) ON CONFLICT (user_id, role_id) DO NOTHING RETURNING id, created_at",
		ur.UserID, ur.RoleID).Scan(&ur.ID, &ur.CreatedAt)
}

func (r *roleRepository) DeleteRole(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM roles WHERE id = $1", id)
	return err
}

func (r *roleRepository) RemoveUserRole(ctx context.Context, userID string, roleID int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2", userID, roleID)
	return err
}
