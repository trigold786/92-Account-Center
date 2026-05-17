package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
)

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) ListRoles(ctx context.Context) ([]model.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRoleRepository) GetRoleByID(ctx context.Context, id int64) (*model.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*model.Role), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockRoleRepository) CreateRole(ctx context.Context, r *model.Role) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *MockRoleRepository) GetRolePermissions(ctx context.Context, roleID int64) ([]model.RolePermission, error) {
	args := m.Called(ctx, roleID)
	return args.Get(0).([]model.RolePermission), args.Error(1)
}

func (m *MockRoleRepository) AddRolePermission(ctx context.Context, rp *model.RolePermission) error {
	args := m.Called(ctx, rp)
	return args.Error(0)
}

func (m *MockRoleRepository) RemoveRolePermission(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) GetUserRoles(ctx context.Context, userID string) ([]model.UserRole, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]model.UserRole), args.Error(1)
}

func (m *MockRoleRepository) SetUserRole(ctx context.Context, ur *model.UserRole) error {
	args := m.Called(ctx, ur)
	return args.Error(0)
}

func (m *MockRoleRepository) RemoveUserRole(ctx context.Context, userID string, roleID int64) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func TestPermissionService_CheckPermission_Allowed(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	userRoles := []model.UserRole{{ID: 1, UserID: "user1", RoleID: 1}}
	permissions := []model.RolePermission{{ID: 1, RoleID: 1, Permission: "config.read"}}

	mockRoleRepo.On("GetUserRoles", mock.Anything, "user1").Return(userRoles, nil)
	mockRoleRepo.On("GetRolePermissions", mock.Anything, int64(1)).Return(permissions, nil)

	allowed, err := svc.CheckPermission(context.Background(), "user1", "config.read")
	assert.NoError(t, err)
	assert.True(t, allowed)
	mockRoleRepo.AssertExpectations(t)
}

func TestPermissionService_CheckPermission_Denied(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	userRoles := []model.UserRole{{ID: 1, UserID: "user1", RoleID: 1}}
	permissions := []model.RolePermission{{ID: 1, RoleID: 1, Permission: "config.read"}}

	mockRoleRepo.On("GetUserRoles", mock.Anything, "user1").Return(userRoles, nil)
	mockRoleRepo.On("GetRolePermissions", mock.Anything, int64(1)).Return(permissions, nil)

	allowed, err := svc.CheckPermission(context.Background(), "user1", "config.edit")
	assert.NoError(t, err)
	assert.False(t, allowed)
	mockRoleRepo.AssertExpectations(t)
}

func TestPermissionService_CheckPermission_NoRoles(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	mockRoleRepo.On("GetUserRoles", mock.Anything, "unknown").Return([]model.UserRole{}, nil)

	allowed, err := svc.CheckPermission(context.Background(), "unknown", "config.read")
	assert.NoError(t, err)
	assert.False(t, allowed)
	mockRoleRepo.AssertExpectations(t)
}

func TestPermissionService_CheckPermission_MultipleRoles(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	userRoles := []model.UserRole{
		{ID: 1, UserID: "user1", RoleID: 1},
		{ID: 2, UserID: "user1", RoleID: 2},
	}
	permsRole1 := []model.RolePermission{}
	permsRole2 := []model.RolePermission{{ID: 3, RoleID: 2, Permission: "config.edit"}}

	mockRoleRepo.On("GetUserRoles", mock.Anything, "user1").Return(userRoles, nil)
	mockRoleRepo.On("GetRolePermissions", mock.Anything, int64(1)).Return(permsRole1, nil)
	mockRoleRepo.On("GetRolePermissions", mock.Anything, int64(2)).Return(permsRole2, nil)

	allowed, err := svc.CheckPermission(context.Background(), "user1", "config.edit")
	assert.NoError(t, err)
	assert.True(t, allowed)
	mockRoleRepo.AssertExpectations(t)
}

func TestPermissionService_CreateRole(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	role := &model.Role{Name: "viewer", Description: "Read-only access"}
	mockRoleRepo.On("CreateRole", mock.Anything, role).Return(nil)
	mockAudit.On("Log", mock.Anything, "CREATE_ROLE", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.CreateRole(context.Background(), role, "admin")
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestPermissionService_ListRoles(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	roles := []model.Role{{ID: 1, Name: "admin"}}
	mockRoleRepo.On("ListRoles", mock.Anything).Return(roles, nil)

	results, err := svc.ListRoles(context.Background())
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	mockRoleRepo.AssertExpectations(t)
}

func TestPermissionService_GetRolePermissions(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	perms := []model.RolePermission{{ID: 1, RoleID: 1, Permission: "config.read"}}
	mockRoleRepo.On("GetRolePermissions", mock.Anything, int64(1)).Return(perms, nil)

	results, err := svc.GetRolePermissions(context.Background(), 1)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	mockRoleRepo.AssertExpectations(t)
}

func TestPermissionService_AddRolePermission(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	rp := &model.RolePermission{RoleID: 1, Permission: "config.edit"}
	mockRoleRepo.On("AddRolePermission", mock.Anything, rp).Return(nil)
	mockAudit.On("Log", mock.Anything, "ADD_PERMISSION", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.AddRolePermission(context.Background(), rp, "admin")
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestPermissionService_SetUserRole(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	ur := &model.UserRole{UserID: "user1", RoleID: 1}
	mockRoleRepo.On("SetUserRole", mock.Anything, ur).Return(nil)
	mockAudit.On("Log", mock.Anything, "SET_USER_ROLE", mock.Anything, "admin", "success", mock.Anything).Return(nil)

	err := svc.SetUserRole(context.Background(), ur, "admin")
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}

func TestPermissionService_GetUserRoles(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	urs := []model.UserRole{{ID: 1, UserID: "user1", RoleID: 1}}
	mockRoleRepo.On("GetUserRoles", mock.Anything, "user1").Return(urs, nil)

	results, err := svc.GetUserRoles(context.Background(), "user1")
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	mockRoleRepo.AssertExpectations(t)
}

func TestPermissionService_RemoveUserRole(t *testing.T) {
	mockRoleRepo := new(MockRoleRepository)
	mockAudit := new(MockAuditSvc)
	svc := NewPermissionService(mockRoleRepo, mockAudit)

	mockRoleRepo.On("RemoveUserRole", mock.Anything, "user1", int64(1)).Return(nil)
	mockAudit.On("Log", mock.Anything, "REMOVE_USER_ROLE", "", "admin", "success", mock.Anything).Return(nil)

	err := svc.RemoveUserRole(context.Background(), "user1", 1, "admin")
	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockAudit.AssertExpectations(t)
}
