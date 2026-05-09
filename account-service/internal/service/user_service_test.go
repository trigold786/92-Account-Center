package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/sunxi/92-Account-Center/account-service/internal/model"
	"github.com/sunxi/92-Account-Center/account-service/internal/repository"
)

// MockUserRepository is a mock implementation of UserRepository for testing.
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	args := m.Called(ctx, user)
	if args.Get(0) != nil {
		user.ID = args.Get(0).(int64)
		user.CreatedAt = args.Get(1).(model.User).CreatedAt
		user.UpdatedAt = args.Get(2).(model.User).UpdatedAt
	}
	return args.Error(3)
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

func TestUserService_Register_Success(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Mock that phone number and account ID don't exist
	mockRepo.On("ExistsByPhoneNumber", mock.Anything, "13800138000").Return(false, nil)
	mockRepo.On("ExistsByAccountID", mock.Anything, "testuser").Return(false, nil)
	
	// Mock successful user creation
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *model.User) {
		return u.PhoneNumber == "13800138000" && 
		       u.AccountID == "testuser" &&
		       len(u.PasswordHash) > 0
	})).Return(nil)

	// Act
	user, err := userService.Register(context.Background(), "13800138000", "testuser", "Password123!", true)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, int64(0), user.ID) // ID will be 0 because we didn't set it in mock
	assert.Equal(t, "13800138000", user.PhoneNumber)
	assert.Equal(t, "testuser", user.AccountID)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_PhoneNumberExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Mock that phone number already exists
	mockRepo.On("ExistsByPhoneNumber", mock.Anything, "13800138000").Return(true, nil)

	// Act
	user, err := userService.Register(context.Background(), "13800138000", "testuser", "Password123!", true)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "phone number already registered", err.Error())
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_AccountIDExists(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Mock that phone number doesn't exist but account ID does
	mockRepo.On("ExistsByPhoneNumber", mock.Anything, "13800138000").Return(false, nil)
	mockRepo.On("ExistsByAccountID", mock.Anything, "testuser").Return(true, nil)

	// Act
	user, err := userService.Register(context.Background(), "13800138000", "testuser", "Password123!", true)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "account ID already taken", err.Error())
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_InvalidPhoneNumber(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Act
	user, err := userService.Register(context.Background(), "invalid", "testuser", "Password123!", true)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid phone number")
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_InvalidAccountID(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Act
	user, err := userService.Register(context.Background(), "13800138000", "123invalid", "Password123!", true)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot start with a number")
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_WeakPassword(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Act
	user, err := userService.Register(context.Background(), "13800138000", "testuser", "weak", true)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password must be between 8 and 20 characters")
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_Register_NoAgreeToTerms(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Act
	user, err := userService.Register(context.Background(), "13800138000", "testuser", "Password123!", false)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, "must agree to terms and conditions", err.Error())
	assert.Nil(t, user)
	mockRepo.AssertExpectations(t)
}

func TestUserService_ValidatePassword(t *testing.T) {
	// Arrange
	mockRepo := new(MockUserRepository)
	userService := NewUserService(mockRepo)

	// Test cases for password validation
	tests := []struct {
		password string
		valid    bool
		errorMsg string
	}{
		{"Abcdefg1!", true, ""}, // Valid password
		{"abcdefg1!", false, "must contain at least one uppercase letter"}, // No uppercase
		{"ABCDEFG1!", false, "must contain at least one lowercase letter"}, // No lowercase
		{"Abcdefg!", false, "must contain at least one digit"}, // No digit
		{"Abcdefg1", false, "must contain at least one special character"}, // No special char
		{"Ab1!", false, "password must be between 8 and 20 characters"}, // Too short
		{"Abcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()_+", false, "password must be between 8 and 20 characters"}, // Too long
	}

	// Act & Assert
	for _, tt := range tests {
		err := userService.ValidatePassword(tt.password)
		if tt.valid {
			assert.NoError(t, err, "Password '%s' should be valid", tt.password)
		} else {
			assert.Error(t, err, "Password '%s' should be invalid", tt.password)
			assert.Contains(t, err.Error(), tt.errorMsg, "Password '%s' error message mismatch", tt.password)
		}
	}
}