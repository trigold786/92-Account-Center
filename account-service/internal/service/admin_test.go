package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

type mockAdminRepo struct {
	mock.Mock
}

func (m *mockAdminRepo) ListUsers(ctx context.Context, page, pageSize int, search string, status string, tier int) ([]model.User, int, error) {
	args := m.Called(ctx, page, pageSize, search, status, tier)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]model.User), args.Int(1), args.Error(2)
}

func (m *mockAdminRepo) GetUserDetail(ctx context.Context, userID int64) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *mockAdminRepo) UpdateUserStatus(ctx context.Context, userID int64, status string, reason string, operator string) error {
	return m.Called(ctx, userID, status, reason, operator).Error(0)
}

func (m *mockAdminRepo) AdjustIdentityTier(ctx context.Context, userID int64, tier int, reason string, operator string) error {
	return m.Called(ctx, userID, tier, reason, operator).Error(0)
}

func (m *mockAdminRepo) InsertAuditLog(ctx context.Context, userID int64, action string, details string, operator string) error {
	return m.Called(ctx, userID, action, details, operator).Error(0)
}

func (m *mockAdminRepo) GetAuditLog(ctx context.Context, userID int64, page, pageSize int) ([]model.AuditLogEntry, int, error) {
	args := m.Called(ctx, userID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]model.AuditLogEntry), args.Int(1), args.Error(2)
}

func (m *mockAdminRepo) EnsureAuditLogTable(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type mockCreditClient struct {
	mock.Mock
}

func (m *mockCreditClient) AdjustCredits(ctx context.Context, userID int64, amount int64, adjustType string) error {
	return m.Called(ctx, userID, amount, adjustType).Error(0)
}

func TestAdmin_ListUsers_Success(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	users := []model.User{
		{ID: 1, PhoneNumber: "13800138000", AccountID: "user1", Status: "active"},
		{ID: 2, PhoneNumber: "13900139000", AccountID: "user2", Status: "active"},
	}
	repo.On("ListUsers", mock.Anything, 1, 20, "", "", 0).Return(users, 2, nil)

	resp, err := svc.ListUsers(context.Background(), &model.AdminUserListRequest{Page: 1, PageSize: 20})
	assert.NoError(t, err)
	assert.Equal(t, 2, resp.Total)
	assert.Len(t, resp.Users, 2)
	repo.AssertExpectations(t)
}

func TestAdmin_ListUsers_Empty(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	repo.On("ListUsers", mock.Anything, 1, 20, "", "", 0).Return([]model.User(nil), 0, nil)

	resp, err := svc.ListUsers(context.Background(), &model.AdminUserListRequest{Page: 1, PageSize: 20})
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Total)
	assert.NotNil(t, resp.Users)
	assert.Len(t, resp.Users, 0)
	repo.AssertExpectations(t)
}

func TestAdmin_ListUsers_Pagination(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	users := []model.User{{ID: 3, AccountID: "user3"}}
	repo.On("ListUsers", mock.Anything, 2, 10, "", "", 0).Return(users, 15, nil)

	resp, err := svc.ListUsers(context.Background(), &model.AdminUserListRequest{Page: 2, PageSize: 10})
	assert.NoError(t, err)
	assert.Equal(t, 15, resp.Total)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
	repo.AssertExpectations(t)
}

func TestAdmin_ListUsers_SearchFilter(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	users := []model.User{{ID: 1, PhoneNumber: "13800138000"}}
	repo.On("ListUsers", mock.Anything, 1, 20, "138", "active", 0).Return(users, 1, nil)

	resp, err := svc.ListUsers(context.Background(), &model.AdminUserListRequest{Page: 1, PageSize: 20, Search: "138", Status: "active"})
	assert.NoError(t, err)
	assert.Equal(t, 1, resp.Total)
	repo.AssertExpectations(t)
}

func TestAdmin_GetUserDetail_Success(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	user := &model.User{ID: 1, AccountID: "user1", Status: "active"}
	repo.On("GetUserDetail", mock.Anything, int64(1)).Return(user, nil)

	result, err := svc.GetUserDetail(context.Background(), 1)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), result.ID)
	repo.AssertExpectations(t)
}

func TestAdmin_GetUserDetail_NotFound(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	repo.On("GetUserDetail", mock.Anything, int64(999)).Return(nil, nil)

	result, err := svc.GetUserDetail(context.Background(), 999)
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, result)
	repo.AssertExpectations(t)
}

func TestAdmin_UpdateUserStatus_Success(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	user := &model.User{ID: 1, AccountID: "user1", Status: "active"}
	repo.On("GetUserDetail", mock.Anything, int64(1)).Return(user, nil)
	repo.On("UpdateUserStatus", mock.Anything, int64(1), "frozen", "violation", "admin_1").Return(nil)

	err := svc.UpdateUserStatus(context.Background(), "admin_1", 1, &model.AdminStatusUpdateRequest{Status: "frozen", Reason: "violation"})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAdmin_UpdateUserStatus_UserNotFound(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	repo.On("GetUserDetail", mock.Anything, int64(999)).Return(nil, nil)

	err := svc.UpdateUserStatus(context.Background(), "admin_1", 999, &model.AdminStatusUpdateRequest{Status: "frozen", Reason: "violation"})
	assert.Equal(t, ErrUserNotFound, err)
	repo.AssertExpectations(t)
}

func TestAdmin_AdjustIdentityTier_Success(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	user := &model.User{ID: 1, IdentityTier: 0}
	repo.On("GetUserDetail", mock.Anything, int64(1)).Return(user, nil)
	repo.On("AdjustIdentityTier", mock.Anything, int64(1), 3, "upgrade", "admin_1").Return(nil)

	err := svc.AdjustIdentityTier(context.Background(), "admin_1", 1, &model.AdminTierUpdateRequest{Tier: 3, Reason: "upgrade"})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
}

func TestAdmin_AdjustIdentityTier_UserNotFound(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	repo.On("GetUserDetail", mock.Anything, int64(999)).Return(nil, nil)

	err := svc.AdjustIdentityTier(context.Background(), "admin_1", 999, &model.AdminTierUpdateRequest{Tier: 3, Reason: "upgrade"})
	assert.Equal(t, ErrUserNotFound, err)
	repo.AssertExpectations(t)
}

func TestAdmin_AdjustCredits_Success(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	user := &model.User{ID: 1}
	repo.On("GetUserDetail", mock.Anything, int64(1)).Return(user, nil)
	credit.On("AdjustCredits", mock.Anything, int64(1), int64(100), "earn").Return(nil)
	repo.On("InsertAuditLog", mock.Anything, int64(1), "adjust_credits", mock.AnythingOfType("string"), "admin_1").Return(nil)

	err := svc.AdjustCredits(context.Background(), "admin_1", 1, &model.AdminCreditAdjustRequest{Amount: 100, Reason: "bonus", Type: "earn"})
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	credit.AssertExpectations(t)
}

func TestAdmin_AdjustCredits_UserNotFound(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	repo.On("GetUserDetail", mock.Anything, int64(999)).Return(nil, nil)

	err := svc.AdjustCredits(context.Background(), "admin_1", 999, &model.AdminCreditAdjustRequest{Amount: 100, Reason: "bonus", Type: "earn"})
	assert.Equal(t, ErrUserNotFound, err)
	repo.AssertExpectations(t)
}

func TestAdmin_AdjustCredits_CreditServiceFailed(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	user := &model.User{ID: 1}
	repo.On("GetUserDetail", mock.Anything, int64(1)).Return(user, nil)
	credit.On("AdjustCredits", mock.Anything, int64(1), int64(100), "earn").Return(errors.New("service down"))

	err := svc.AdjustCredits(context.Background(), "admin_1", 1, &model.AdminCreditAdjustRequest{Amount: 100, Reason: "bonus", Type: "earn"})
	assert.Equal(t, ErrCreditAdjustFailed, err)
	repo.AssertExpectations(t)
	credit.AssertExpectations(t)
}

func TestAdmin_GetAuditLog_Success(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	entries := []model.AuditLogEntry{
		{ID: 1, UserID: 1, Action: "update_status", Operator: "admin_1"},
		{ID: 2, UserID: 1, Action: "adjust_tier", Operator: "admin_1"},
	}
	repo.On("GetAuditLog", mock.Anything, int64(1), 1, 20).Return(entries, 2, nil)

	result, total, err := svc.GetAuditLog(context.Background(), 1, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

func TestAdmin_GetAuditLog_Empty(t *testing.T) {
	repo := new(mockAdminRepo)
	credit := new(mockCreditClient)
	svc := NewAdminService(repo, credit)

	repo.On("GetAuditLog", mock.Anything, int64(1), 1, 20).Return([]model.AuditLogEntry(nil), 0, nil)

	result, total, err := svc.GetAuditLog(context.Background(), 1, 1, 20)
	assert.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
	repo.AssertExpectations(t)
}
