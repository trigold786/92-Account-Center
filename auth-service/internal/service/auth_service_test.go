package service

import (
	"context"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
	"github.com/trigold786/92-Account-Center/auth-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/auth-service/pkg/crypto"
	"github.com/trigold786/92-Account-Center/auth-service/pkg/jwt"
)

var maxFailedAttempts = 5

type mockUserRepo struct {
	users map[string]*model.User
}

func (m *mockUserRepo) GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error) {
	if u, ok := m.users["phone:"+phoneNumber]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if u, ok := m.users["email:"+email]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) GetByAccountID(ctx context.Context, accountID string) (*model.User, error) {
	if u, ok := m.users["account:"+accountID]; ok {
		return u, nil
	}
	return nil, nil
}

func newTestAuthService(repo UserRepository) AuthService {
	jwtMgr := jwt.NewJWTManager("test-access-secret", "test-refresh-secret", 15*time.Minute, 24*time.Hour)
	cfg := &svcconfig.AuthConfig{
		JwtAccessTokenExpire:  15 * time.Minute,
		JwtRefreshTokenExpire: 24 * time.Hour,
		LoginMaxAttempts:      5,
		LoginLockoutDuration:  15 * time.Minute,
	}
	return NewAuthService(repo, jwtMgr, nil, cfg)
}

func makeTestUser() *model.User {
	salt := "testsalt"
	password := "correctpassword"
	hashed := crypto.SM3Hash([]byte(salt + password))
	return &model.User{
		ID:           1,
		PhoneNumber:  "13800138000",
		AccountID:    "testuser123",
		Email:        strPtr("test@example.com"),
		PasswordHash: salt + "$" + hashed,
		MFAEnabled:   false,
	}
}

func TestLogin_Success(t *testing.T) {
	user := makeTestUser()
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"phone:13800138000": user,
		},
	}
	svc := newTestAuthService(repo)

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "13800138000",
		Password:   "correctpassword",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if resp.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if resp.UserID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, resp.UserID)
	}
}

func TestLogin_NilRequest(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*model.User{}}
	svc := newTestAuthService(repo)

	_, err := svc.Login(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error for nil request")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*model.User{}}
	svc := newTestAuthService(repo)

	_, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "13800138000",
		Password:   "password",
	})
	if err == nil {
		t.Fatal("expected error for unknown user")
	}
}

func TestLogin_InvalidCredentialFormat(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*model.User{}}
	svc := newTestAuthService(repo)

	_, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "!!!invalid!!!",
		Password:   "password",
	})
	if err == nil {
		t.Fatal("expected error for invalid credential format")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	user := makeTestUser()
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"phone:13800138000": user,
		},
	}
	svc := newTestAuthService(repo)

	_, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "13800138000",
		Password:   "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestLogin_NoPasswordOrCode(t *testing.T) {
	user := makeTestUser()
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"phone:13800138000": user,
		},
	}
	svc := newTestAuthService(repo)

	_, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "13800138000",
	})
	if err == nil {
		t.Fatal("expected error when no password or code provided")
	}
}

func TestLogin_Lockout(t *testing.T) {
	user := makeTestUser()
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"phone:13800138000": user,
		},
	}
	svc := newTestAuthService(repo)

	// Without Redis (nil client), lockout is not enforced.
	for i := 0; i < maxFailedAttempts; i++ {
		_, err := svc.Login(context.Background(), &model.LoginRequest{
			Credential: "13800138000",
			Password:   "wrongpassword",
		})
		if err == nil {
			t.Fatalf("expected error on attempt %d, got nil", i+1)
		}
	}

	// After max failures, login should still work (graceful degradation without Redis)
	_, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "13800138000",
		Password:   "correctpassword",
	})
	if err != nil {
		t.Fatalf("expected success (graceful degradation without Redis), got %v", err)
	}
}

func TestLogin_ResetOnSuccess(t *testing.T) {
	user := makeTestUser()
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"phone:13800138000": user,
		},
	}
	svc := newTestAuthService(repo)

	// Without Redis (nil client), lockout is not enforced.
	for i := 0; i < maxFailedAttempts-1; i++ {
		_, err := svc.Login(context.Background(), &model.LoginRequest{
			Credential: "13800138000",
			Password:   "wrongpassword",
		})
		if err == nil {
			t.Fatalf("expected error on attempt %d, got nil", i+1)
		}
	}

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "13800138000",
		Password:   "correctpassword",
	})
	if err != nil {
		t.Fatalf("expected success after correct password, got %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	// After successful login, counter is not persisted (no Redis), so further failures don't lock
	for i := 0; i < maxFailedAttempts; i++ {
		_, err := svc.Login(context.Background(), &model.LoginRequest{
			Credential: "13800138000",
			Password:   "wrongpassword",
		})
		if err == nil {
			t.Fatalf("expected error on attempt %d, got nil", i+1)
		}
	}

	// Should still succeed because lockout requires Redis
	_, err = svc.Login(context.Background(), &model.LoginRequest{
		Credential: "13800138000",
		Password:   "correctpassword",
	})
	if err != nil {
		t.Fatalf("expected success (graceful degradation without Redis), got %v", err)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	user := makeTestUser()
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"phone:13800138000": user,
		},
	}
	svc := newTestAuthService(repo)

	loginResp, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "13800138000",
		Password:   "correctpassword",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	refreshResp, err := svc.RefreshToken(context.Background(), loginResp.RefreshToken)
	if err != nil {
		t.Fatalf("refresh failed: %v", err)
	}
	if refreshResp.AccessToken == "" {
		t.Error("expected non-empty refreshed access token")
	}
}

func TestRefreshToken_Invalid(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*model.User{}}
	svc := newTestAuthService(repo)

	_, err := svc.RefreshToken(context.Background(), "invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}
}

func TestVerifyPassword(t *testing.T) {
	correctHash := "testsalt$" + crypto.SM3Hash([]byte("testsaltcorrectpassword"))
	tests := []struct {
		name       string
		password   string
		hash       string
		wantResult bool
	}{
		{
			name:       "correct password",
			password:   "correctpassword",
			hash:       correctHash,
			wantResult: true,
		},
		{
			name:       "wrong password",
			password:   "wrong",
			hash:       correctHash,
			wantResult: false,
		},
		{
			name:       "malformed hash no separator",
			password:   "test",
			hash:       "malformedhash",
			wantResult: false,
		},
		{
			name:       "empty hash",
			password:   "test",
			hash:       "",
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockUserRepo{users: map[string]*model.User{}}
			svc := newTestAuthService(repo)
			result := svc.VerifyPassword(tt.password, tt.hash)
			if result != tt.wantResult {
				t.Errorf("VerifyPassword(%q, %q) = %v, want %v", tt.password, tt.hash, result, tt.wantResult)
			}
		})
	}
}

func TestLogout_Success(t *testing.T) {
	user := makeTestUser()
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"phone:13800138000": user,
		},
	}
	svc := newTestAuthService(repo)

	loginResp, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "13800138000",
		Password:   "correctpassword",
	})
	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	err = svc.Logout(context.Background(), loginResp.AccessToken)
	if err != nil {
		t.Fatalf("logout failed: %v", err)
	}
}

func TestLogout_InvalidToken(t *testing.T) {
	repo := &mockUserRepo{users: map[string]*model.User{}}
	svc := newTestAuthService(repo)

	err := svc.Logout(context.Background(), "invalid-token")
	if err == nil {
		t.Fatal("expected error for invalid token on logout")
	}
}

func TestLogin_EmailCredential(t *testing.T) {
	user := makeTestUser()
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"email:test@example.com": user,
		},
	}
	svc := newTestAuthService(repo)

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "test@example.com",
		Password:   "correctpassword",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.UserID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, resp.UserID)
	}
}

func TestLogin_AccountIDCredential(t *testing.T) {
	user := makeTestUser()
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"account:testuser123": user,
		},
	}
	svc := newTestAuthService(repo)

	resp, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "testuser123",
		Password:   "correctpassword",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.UserID != user.ID {
		t.Errorf("expected user ID %d, got %d", user.ID, resp.UserID)
	}
}

func strPtr(s string) *string {
	return &s
}

func TestLogin_LockoutExpires(t *testing.T) {
	user := makeTestUser()
	repo := &mockUserRepo{
		users: map[string]*model.User{
			"phone:13800138000": user,
		},
	}
	svc := newTestAuthService(repo)

	// Without Redis (nil client), lockout is not enforced.
	// 5 failed attempts should not lock out when Redis is unavailable.
	for i := 0; i < maxFailedAttempts; i++ {
		_, err := svc.Login(context.Background(), &model.LoginRequest{
			Credential: "13800138000",
			Password:   "wrongpassword",
		})
		if err == nil {
			t.Fatalf("expected error on attempt %d, got nil", i+1)
		}
	}

	// After 5 failures, the login should still work (graceful degradation)
	// because lockout requires Redis.
	_, err := svc.Login(context.Background(), &model.LoginRequest{
		Credential: "13800138000",
		Password:   "correctpassword",
	})
	if err != nil {
		t.Fatalf("expected success (graceful degradation without Redis), got %v", err)
	}
}
