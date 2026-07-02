package service

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelWarn}))
}

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) Create(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	args := m.Called(ctx, phoneNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) GetByAccountID(ctx context.Context, accountID string) (*model.User, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) GetByID(ctx context.Context, userID string) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}
func (m *mockUserRepo) Update(ctx context.Context, user *model.User) error {
	return m.Called(ctx, user).Error(0)
}
func (m *mockUserRepo) ExistsByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error) {
	args := m.Called(ctx, phoneNumber)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepo) ExistsByAccountID(ctx context.Context, accountID string) (bool, error) {
	args := m.Called(ctx, accountID)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}
func (m *mockUserRepo) PermanentDelete(ctx context.Context, userID int64) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockUserRepo) AnonymizeUser(ctx context.Context, userID int64) error {
	return m.Called(ctx, userID).Error(0)
}
func (m *mockUserRepo) GetExpiredDeletions(ctx context.Context) ([]model.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.User), args.Error(1)
}
func (m *mockUserRepo) UpdateEmail(ctx context.Context, id int64, email string) error {
	return m.Called(ctx, id, email).Error(0)
}
func (m *mockUserRepo) UpdatePhone(ctx context.Context, id int64, phone string) error {
	return m.Called(ctx, id, phone).Error(0)
}
func (m *mockUserRepo) UpdateIdentityTier(ctx context.Context, userID int64, tier int) error {
	return m.Called(ctx, userID, tier).Error(0)
}
func (m *mockUserRepo) GetIdentityTier(ctx context.Context, userID int64) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}
func (m *mockUserRepo) WriteDeletionAudit(ctx context.Context, userID int64, details map[string]interface{}) error {
	return m.Called(ctx, userID, details).Error(0)
}
func (m *mockUserRepo) AnonymizeEnterprisePII(ctx context.Context, userID int64) error {
	return m.Called(ctx, userID).Error(0)
}

type mockEntitlementRepo struct {
	mock.Mock
}

func (m *mockEntitlementRepo) Create(ctx context.Context, e *model.Entitlement) error {
	return m.Called(ctx, e).Error(0)
}
func (m *mockEntitlementRepo) GetByUserID(ctx context.Context, userID int64) ([]model.Entitlement, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Entitlement), args.Error(1)
}
func (m *mockEntitlementRepo) GetByUserAndFeature(ctx context.Context, userID int64, featureCode string) (*model.Entitlement, error) {
	args := m.Called(ctx, userID, featureCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Entitlement), args.Error(1)
}
func (m *mockEntitlementRepo) UpdateQuota(ctx context.Context, id int64, usedQuota int) error {
	return m.Called(ctx, id, usedQuota).Error(0)
}
func (m *mockEntitlementRepo) UpdateTotalQuota(ctx context.Context, id int64, totalQuota int) error {
	return m.Called(ctx, id, totalQuota).Error(0)
}
func (m *mockEntitlementRepo) DeleteByUserID(ctx context.Context, userID int64) error {
	return m.Called(ctx, userID).Error(0)
}

func TestProcessExpiredDeletions_NoExpired(t *testing.T) {
	userRepo := new(mockUserRepo)
	entRepo := new(mockEntitlementRepo)

	userRepo.On("GetExpiredDeletions", mock.Anything).Return([]model.User(nil), nil)

	svc := NewDeletionService(
		userRepo,
		entRepo,
		(*redis.Client)(nil),
		testLogger(),
	)

	count, err := svc.ProcessExpiredDeletions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	userRepo.AssertExpectations(t)
}

func TestProcessExpiredDeletions_WithExpired(t *testing.T) {
	userRepo := new(mockUserRepo)
	entRepo := new(mockEntitlementRepo)

	users := []model.User{{ID: 1}, {ID: 2}}
	userRepo.On("GetExpiredDeletions", mock.Anything).Return(users, nil)
	userRepo.On("AnonymizeUser", mock.Anything, int64(1)).Return(nil)
	userRepo.On("AnonymizeUser", mock.Anything, int64(2)).Return(nil)
	userRepo.On("AnonymizeEnterprisePII", mock.Anything, int64(1)).Return(nil)
	userRepo.On("AnonymizeEnterprisePII", mock.Anything, int64(2)).Return(nil)
	entRepo.On("DeleteByUserID", mock.Anything, int64(1)).Return(nil)
	entRepo.On("DeleteByUserID", mock.Anything, int64(2)).Return(nil)
	userRepo.On("WriteDeletionAudit", mock.Anything, int64(1), mock.Anything).Return(nil)
	userRepo.On("WriteDeletionAudit", mock.Anything, int64(2), mock.Anything).Return(nil)

	svc := NewDeletionService(
		userRepo,
		entRepo,
		(*redis.Client)(nil),
		testLogger(),
	)

	count, err := svc.ProcessExpiredDeletions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	userRepo.AssertExpectations(t)
	entRepo.AssertExpectations(t)
}

func TestProcessExpiredDeletions_PartialFailure(t *testing.T) {
	userRepo := new(mockUserRepo)
	entRepo := new(mockEntitlementRepo)

	users := []model.User{{ID: 1}, {ID: 2}}
	userRepo.On("GetExpiredDeletions", mock.Anything).Return(users, nil)
	userRepo.On("AnonymizeUser", mock.Anything, int64(1)).Return(errors.New("db error"))
	userRepo.On("AnonymizeUser", mock.Anything, int64(2)).Return(nil)
	userRepo.On("AnonymizeEnterprisePII", mock.Anything, int64(2)).Return(nil)
	entRepo.On("DeleteByUserID", mock.Anything, int64(2)).Return(nil)
	userRepo.On("WriteDeletionAudit", mock.Anything, int64(2), mock.Anything).Return(nil)

	svc := NewDeletionService(
		userRepo,
		entRepo,
		(*redis.Client)(nil),
		testLogger(),
	)

	count, err := svc.ProcessExpiredDeletions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	userRepo.AssertExpectations(t)
	entRepo.AssertExpectations(t)
}

func TestProcessExpiredDeletions_QueryError(t *testing.T) {
	userRepo := new(mockUserRepo)
	entRepo := new(mockEntitlementRepo)

	userRepo.On("GetExpiredDeletions", mock.Anything).Return([]model.User(nil), errors.New("connection lost"))

	svc := NewDeletionService(
		userRepo,
		entRepo,
		(*redis.Client)(nil),
		testLogger(),
	)

	count, err := svc.ProcessExpiredDeletions(context.Background())
	assert.Error(t, err)
	assert.Equal(t, 0, count)
}

func TestDeletionService_Interface(t *testing.T) {
	var _ repository.UserRepository = (*mockUserRepo)(nil)
	var _ repository.EntitlementRepository = (*mockEntitlementRepo)(nil)
}

func TestProcessExpiredDeletions_WritesAuditLog(t *testing.T) {
	userRepo := new(mockUserRepo)
	entRepo := new(mockEntitlementRepo)
	users := []model.User{{ID: 42}}
	userRepo.On("GetExpiredDeletions", mock.Anything).Return(users, nil)
	userRepo.On("AnonymizeUser", mock.Anything, int64(42)).Return(nil)
	userRepo.On("AnonymizeEnterprisePII", mock.Anything, int64(42)).Return(nil)
	entRepo.On("DeleteByUserID", mock.Anything, int64(42)).Return(nil)
	userRepo.On("WriteDeletionAudit", mock.Anything, int64(42), mock.Anything).Return(nil)

	svc := NewDeletionService(userRepo, entRepo, (*redis.Client)(nil), testLogger())
	count, err := svc.ProcessExpiredDeletions(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
	userRepo.AssertExpectations(t)
}
