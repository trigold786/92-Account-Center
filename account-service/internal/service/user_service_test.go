package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
)

type MockUserRepository struct {
	users         map[string]*model.User
	phoneIndex    map[string]int64
	accountIDIndex map[string]int64
	nextID        int64
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users:         make(map[string]*model.User),
		phoneIndex:    make(map[string]int64),
		accountIDIndex: make(map[string]int64),
		nextID:        1,
	}
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) error {
	user.ID = m.nextID
	m.nextID++
	m.users[int64ToString(user.ID)] = user
	m.phoneIndex[user.PhoneNumber] = user.ID
	m.accountIDIndex[user.AccountID] = user.ID
	return nil
}

func (m *MockUserRepository) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	if id, ok := m.phoneIndex[phoneNumber]; ok {
		return m.users[int64ToString(id)], nil
	}
	return nil, nil
}

func (m *MockUserRepository) GetByAccountID(ctx context.Context, accountID string) (*model.User, error) {
	if id, ok := m.accountIDIndex[accountID]; ok {
		return m.users[int64ToString(id)], nil
	}
	return nil, nil
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	return nil, nil
}

func (m *MockUserRepository) GetByID(ctx context.Context, userID string) (*model.User, error) {
	if user, ok := m.users[userID]; ok {
		return user, nil
	}
	return nil, nil
}

func (m *MockUserRepository) Update(ctx context.Context, user *model.User) error {
	m.users[int64ToString(user.ID)] = user
	return nil
}

func (m *MockUserRepository) ExistsByPhoneNumber(ctx context.Context, phoneNumber string) (bool, error) {
	_, ok := m.phoneIndex[phoneNumber]
	return ok, nil
}

func (m *MockUserRepository) ExistsByAccountID(ctx context.Context, accountID string) (bool, error) {
	_, ok := m.accountIDIndex[accountID]
	return ok, nil
}

func (m *MockUserRepository) PermanentDelete(ctx context.Context, userID int64) error {
	delete(m.users, int64ToString(userID))
	return nil
}

func (m *MockUserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	for _, u := range m.users {
		if u.Email == email {
			return true, nil
		}
	}
	return false, nil
}

func (m *MockUserRepository) UpdateEmail(ctx context.Context, id int64, email string) error {
	if u, ok := m.users[int64ToString(id)]; ok {
		u.Email = email
		return nil
	}
	return nil
}

func (m *MockUserRepository) UpdatePhone(ctx context.Context, id int64, phone string) error {
	if u, ok := m.users[int64ToString(id)]; ok {
		delete(m.phoneIndex, u.PhoneNumber)
		u.PhoneNumber = phone
		m.phoneIndex[phone] = id
		return nil
	}
	return nil
}

func (m *MockUserRepository) UpdateIdentityTier(ctx context.Context, userID int64, tier int) error {
	if u, ok := m.users[int64ToString(userID)]; ok {
		u.IdentityTier = tier
		return nil
	}
	return nil
}

func (m *MockUserRepository) GetIdentityTier(ctx context.Context, userID int64) (int, error) {
	if u, ok := m.users[int64ToString(userID)]; ok {
		return u.IdentityTier, nil
	}
	return 0, nil
}

func int64ToString(n int64) string {
	return string(rune(n))
}

func TestUserService_Register_Success(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewUserService(repo, nil)

	user, err := svc.Register(context.Background(), "13800138000", "testuser", "Password123!", true)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if user.PhoneNumber != "13800138000" {
		t.Errorf("expected phone 13800138000, got %s", user.PhoneNumber)
	}
	if user.AccountID != "testuser" {
		t.Errorf("expected accountID testuser, got %s", user.AccountID)
	}
}

func TestUserService_Register_PhoneExists(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewUserService(repo, nil)

	_, _ = svc.Register(context.Background(), "13800138000", "user1", "Password123!", true)
	_, err := svc.Register(context.Background(), "13800138000", "user2", "Password123!", true)
	if err == nil {
		t.Fatal("expected error for duplicate phone")
	}
}

func TestUserService_Register_AccountIDExists(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewUserService(repo, nil)

	_, _ = svc.Register(context.Background(), "13800138000", "testuser", "Password123!", true)
	_, err := svc.Register(context.Background(), "13900139000", "testuser", "Password123!", true)
	if err == nil {
		t.Fatal("expected error for duplicate account ID")
	}
}

func TestUserService_Register_InvalidPhone(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewUserService(repo, nil)

	_, err := svc.Register(context.Background(), "invalid", "testuser", "Password123!", true)
	if err == nil {
		t.Fatal("expected error for invalid phone")
	}
}

func TestUserService_Register_AccountIDStartsWithDigit(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewUserService(repo, nil)

	_, err := svc.Register(context.Background(), "13800138000", "123invalid", "Password123!", true)
	if err == nil {
		t.Fatal("expected error for account ID starting with digit")
	}
}

func TestUserService_Register_WeakPassword(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewUserService(repo, nil)

	_, err := svc.Register(context.Background(), "13800138000", "testuser", "weak", true)
	if err == nil {
		t.Fatal("expected error for weak password")
	}
}

func TestUserService_Register_NoAgreement(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewUserService(repo, nil)

	_, err := svc.Register(context.Background(), "13800138000", "testuser", "Password123!", false)
	if err == nil {
		t.Fatal("expected error when not agreeing to terms")
	}
}

func TestUserService_ValidatePassword(t *testing.T) {
	repo := NewMockUserRepository()
	svc := NewUserService(repo, nil)

	tests := []struct {
		password string
		valid    bool
	}{
		{"Abcdefg1!", true},
		{"abcdefg1!", false},
		{"ABCDEFG1!", false},
		{"Abcdefg!", false},
		{"Abcdefg1", false},
		{"Ab1!", false},
	}

	for _, tt := range tests {
		err := svc.ValidatePassword(tt.password)
		if tt.valid && err != nil {
			t.Errorf("password '%s' should be valid: %v", tt.password, err)
		}
		if !tt.valid && err == nil {
			t.Errorf("password '%s' should be invalid", tt.password)
		}
	}
}
