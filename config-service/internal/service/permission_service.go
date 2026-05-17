package service

import (
	"context"
	"fmt"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
	"github.com/trigold786/92-Account-Center/config-service/internal/repository"
)

type PermissionService interface {
	ListRoles(ctx context.Context) ([]model.Role, error)
	CreateRole(ctx context.Context, role *model.Role, operator string) error
	GetRolePermissions(ctx context.Context, roleID int64) ([]model.RolePermission, error)
	AddRolePermission(ctx context.Context, rp *model.RolePermission, operator string) error
	RemoveRolePermission(ctx context.Context, id int64, operator string) error
	GetUserRoles(ctx context.Context, userID string) ([]model.UserRole, error)
	SetUserRole(ctx context.Context, ur *model.UserRole, operator string) error
	RemoveUserRole(ctx context.Context, userID string, roleID int64, operator string) error
	CheckPermission(ctx context.Context, userID, permission string) (bool, error)
}

type permissionService struct {
	roleRepo repository.RoleRepository
	auditSvc AuditService
}

func NewPermissionService(roleRepo repository.RoleRepository, auditSvc AuditService) PermissionService {
	return &permissionService{
		roleRepo: roleRepo,
		auditSvc: auditSvc,
	}
}

func (s *permissionService) ListRoles(ctx context.Context) ([]model.Role, error) {
	return s.roleRepo.ListRoles(ctx)
}

func (s *permissionService) CreateRole(ctx context.Context, role *model.Role, operator string) error {
	if err := s.roleRepo.CreateRole(ctx, role); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "CREATE_ROLE", fmt.Sprintf("roles:%d", role.ID), operator, "success",
		fmt.Sprintf("Created role: %s", role.Name))
	return nil
}

func (s *permissionService) GetRolePermissions(ctx context.Context, roleID int64) ([]model.RolePermission, error) {
	return s.roleRepo.GetRolePermissions(ctx, roleID)
}

func (s *permissionService) AddRolePermission(ctx context.Context, rp *model.RolePermission, operator string) error {
	if err := s.roleRepo.AddRolePermission(ctx, rp); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "ADD_PERMISSION", fmt.Sprintf("role_permissions:%d", rp.ID), operator, "success",
		fmt.Sprintf("Added permission %s to role %d", rp.Permission, rp.RoleID))
	return nil
}

func (s *permissionService) RemoveRolePermission(ctx context.Context, id int64, operator string) error {
	if err := s.roleRepo.RemoveRolePermission(ctx, id); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "REMOVE_PERMISSION", "", operator, "success", "")
	return nil
}

func (s *permissionService) GetUserRoles(ctx context.Context, userID string) ([]model.UserRole, error) {
	return s.roleRepo.GetUserRoles(ctx, userID)
}

func (s *permissionService) SetUserRole(ctx context.Context, ur *model.UserRole, operator string) error {
	if err := s.roleRepo.SetUserRole(ctx, ur); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "SET_USER_ROLE", fmt.Sprintf("user_roles:%s->%d", ur.UserID, ur.RoleID), operator, "success", "")
	return nil
}

func (s *permissionService) RemoveUserRole(ctx context.Context, userID string, roleID int64, operator string) error {
	if err := s.roleRepo.RemoveUserRole(ctx, userID, roleID); err != nil {
		return err
	}
	s.auditSvc.Log(ctx, "REMOVE_USER_ROLE", "", operator, "success", "")
	return nil
}

func (s *permissionService) CheckPermission(ctx context.Context, userID, permission string) (bool, error) {
	userRoles, err := s.roleRepo.GetUserRoles(ctx, userID)
	if err != nil {
		return false, err
	}
	for _, ur := range userRoles {
		perms, err := s.roleRepo.GetRolePermissions(ctx, ur.RoleID)
		if err != nil {
			return false, err
		}
		for _, p := range perms {
			if p.Permission == permission {
				return true, nil
			}
		}
	}
	return false, nil
}
