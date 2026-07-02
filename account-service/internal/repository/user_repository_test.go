package repository

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	args := m.Called(ctx, phoneNumber)
	if args.Get(0) != nil {
		return args.Get(0).(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByAccountID(ctx context.Context, accountID string) (*model.User, error) {
	args := m.Called(ctx, accountID)
	if args.Get(0) != nil {
		return args.Get(0).(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) ExistsByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error) {
	args := m.Called(ctx, phoneNumber)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) ExistsByAccountID(ctx context.Context, accountID string) (bool, error) {
	args := m.Called(ctx, accountID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) != nil {
		return args.Get(0).(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*model.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) != nil {
		return args.Get(0).(*model.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	args := m.Called(ctx, email)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) PermanentDelete(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateEmail(ctx context.Context, id int64, email string) error {
	args := m.Called(ctx, id, email)
	return args.Error(0)
}

func (m *MockUserRepository) UpdatePhone(ctx context.Context, id int64, phone string) error {
	args := m.Called(ctx, id, phone)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateIdentityTier(ctx context.Context, userID int64, tier int) error {
	args := m.Called(ctx, userID, tier)
	return args.Error(0)
}

func (m *MockUserRepository) GetIdentityTier(ctx context.Context, userID int64) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) WriteDeletionAudit(ctx context.Context, userID int64, details map[string]interface{}) error {
	args := m.Called(ctx, userID, details)
	return args.Error(0)
}

func (m *MockUserRepository) AnonymizeEnterprisePII(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestUserRepository_Create(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	user := &model.User{
		PhoneNumber:  "13800138000",
		AccountID:    "testuser",
		PasswordHash: "hashedpassword",
	}

	// Act
	mockRepo.On("Create", mock.Anything, user).Return(nil)
	err := mockRepo.Create(context.Background(), user)

	// Assert
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_GetByPhoneNumber(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	expectedUser := &model.User{
		ID:           1,
		PhoneNumber:  "13800138000",
		AccountID:    "testuser",
		PasswordHash: "hashedpassword",
	}

	// Act
	mockRepo.On("GetByPhoneNumber", mock.Anything, "13800138000").Return(expectedUser, nil)
	user, err := mockRepo.GetByPhoneNumber(context.Background(), "13800138000")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "13800138000", user.PhoneNumber)
	assert.Equal(t, "testuser", user.AccountID)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_GetByPhoneNumber_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)

	mockRepo.On("GetByPhoneNumber", mock.Anything, "13800138000").Return(nil, nil)
	user, err := mockRepo.GetByPhoneNumber(context.Background(), "13800138000")

	assert.NoError(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_GetByAccountID(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	expectedUser := &model.User{
		ID:           1,
		PhoneNumber:  "13800138000",
		AccountID:    "testuser",
		PasswordHash: "hashedpassword",
	}

	// Act
	mockRepo.On("GetByAccountID", mock.Anything, "testuser").Return(expectedUser, nil)
	user, err := mockRepo.GetByAccountID(context.Background(), "testuser")

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(1), user.ID)
	assert.Equal(t, "13800138000", user.PhoneNumber)
	assert.Equal(t, "testuser", user.AccountID)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_GetByAccountID_NotFound(t *testing.T) {
	mockRepo := new(MockUserRepository)

	mockRepo.On("GetByAccountID", mock.Anything, "testuser").Return(nil, nil)
	user, err := mockRepo.GetByAccountID(context.Background(), "testuser")

	assert.NoError(t, err)
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_ExistsByPhoneNumber(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)

	// Act
	mockRepo.On("ExistsByPhoneNumber", mock.Anything, "13800138000").Return(true, nil)
	exists, err := mockRepo.ExistsByPhoneNumber(context.Background(), "13800138000")

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_ExistsByPhoneNumber_NotExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)

	// Act
	mockRepo.On("ExistsByPhoneNumber", mock.Anything, "13800138000").Return(false, nil)
	exists, err := mockRepo.ExistsByPhoneNumber(context.Background(), "13800138000")

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_ExistsByAccountID(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)

	// Act
	mockRepo.On("ExistsByAccountID", mock.Anything, "testuser").Return(true, nil)
	exists, err := mockRepo.ExistsByAccountID(context.Background(), "testuser")

	// Assert
	assert.NoError(t, err)
	assert.True(t, exists)
	mockRepo.AssertExpectations(t)
}

func TestUserRepository_ExistsByAccountID_NotExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)

	// Act
	mockRepo.On("ExistsByAccountID", mock.Anything, "testuser").Return(false, nil)
	exists, err := mockRepo.ExistsByAccountID(context.Background(), "testuser")

	// Assert
	assert.NoError(t, err)
	assert.False(t, exists)
	mockRepo.AssertExpectations(t)
}
