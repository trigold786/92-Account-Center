package jwt

import (
	"testing"
	"time"
)

func TestGenerateTokenPair(t *testing.T) {
	mgr := NewJWTManager("access-secret", "refresh-secret")

	accessToken, refreshToken, err := mgr.GenerateTokenPair(1, "account123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if accessToken == "" {
		t.Error("expected non-empty access token")
	}
	if refreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
	if accessToken == refreshToken {
		t.Error("access and refresh tokens should be different")
	}
}

func TestValidateToken_AccessToken(t *testing.T) {
	mgr := NewJWTManager("access-secret", "refresh-secret")

	accessToken, _, err := mgr.GenerateTokenPair(42, "acct42")
	if err != nil {
		t.Fatalf("generate token pair failed: %v", err)
	}

	claims, err := mgr.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("expected no error validating access token, got %v", err)
	}
	if claims.UserID != 42 {
		t.Errorf("expected user ID 42, got %d", claims.UserID)
	}
	if claims.AccountID != "acct42" {
		t.Errorf("expected account ID acct42, got %s", claims.AccountID)
	}
}

func TestValidateToken_RefreshToken(t *testing.T) {
	mgr := NewJWTManager("access-secret", "refresh-secret")

	_, refreshToken, err := mgr.GenerateTokenPair(10, "acct10")
	if err != nil {
		t.Fatalf("generate token pair failed: %v", err)
	}

	claims, err := mgr.ValidateToken(refreshToken)
	if err != nil {
		t.Fatalf("expected no error validating refresh token, got %v", err)
	}
	if claims.UserID != 10 {
		t.Errorf("expected user ID 10, got %d", claims.UserID)
	}
}

func TestValidateToken_Invalid(t *testing.T) {
	mgr := NewJWTManager("access-secret", "refresh-secret")

	_, err := mgr.ValidateToken("this.is.invalid")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	mgr1 := NewJWTManager("secret1", "secret1")
	mgr2 := NewJWTManager("secret2", "secret2")

	token, _, _ := mgr1.GenerateTokenPair(1, "acct1")

	_, err := mgr2.ValidateToken(token)
	if err == nil {
		t.Error("expected error when validating token with wrong secret")
	}
}

func TestRefreshAccessToken(t *testing.T) {
	mgr := NewJWTManager("access-secret", "refresh-secret")

	_, refreshToken, err := mgr.GenerateTokenPair(5, "acct5")
	if err != nil {
		t.Fatalf("generate token pair failed: %v", err)
	}

	newAccessToken, err := mgr.RefreshAccessToken(refreshToken)
	if err != nil {
		t.Fatalf("expected no error refreshing token, got %v", err)
	}
	if newAccessToken == "" {
		t.Error("expected non-empty new access token")
	}

	claims, err := mgr.ValidateToken(newAccessToken)
	if err != nil {
		t.Fatalf("expected no error validating refreshed access token, got %v", err)
	}
	if claims.UserID != 5 {
		t.Errorf("expected user ID 5, got %d", claims.UserID)
	}
}

func TestRefreshAccessToken_InvalidToken(t *testing.T) {
	mgr := NewJWTManager("access-secret", "refresh-secret")

	_, err := mgr.RefreshAccessToken("invalid-token")
	if err == nil {
		t.Error("expected error for invalid refresh token")
	}
}

func TestRefreshAccessToken_WithAccessTokenFails(t *testing.T) {
	mgr := NewJWTManager("access-secret", "refresh-secret")

	accessToken, _, _ := mgr.GenerateTokenPair(1, "acct1")

	_, err := mgr.RefreshAccessToken(accessToken)
	if err == nil {
		t.Error("expected error when using access token as refresh token")
	}
}

func TestTokenExpiry(t *testing.T) {
	mgr := NewJWTManager("access-secret", "refresh-secret")

	accessToken, _, _ := mgr.GenerateTokenPair(1, "acct1")

	claims, err := mgr.ValidateToken(accessToken)
	if err != nil {
		t.Fatalf("validate token failed: %v", err)
	}

	if claims.ExpiresAt == nil {
		t.Error("expected ExpiresAt to be set")
	}
	if claims.IssuedAt == nil {
		t.Error("expected IssuedAt to be set")
	}

	expiry := claims.ExpiresAt.Sub(claims.IssuedAt.Time)
	expectedExpiry := 24 * time.Hour
	if expiry != expectedExpiry {
		t.Errorf("expected expiry %v, got %v", expectedExpiry, expiry)
	}
}

func TestMultipleTokenPairs(t *testing.T) {
	mgr := NewJWTManager("access-secret", "refresh-secret")

	tests := []struct {
		userID    int64
		accountID string
	}{
		{1, "account1"},
		{2, "account2"},
		{999, "account999"},
	}

	for _, tt := range tests {
		accessToken, _, err := mgr.GenerateTokenPair(tt.userID, tt.accountID)
		if err != nil {
			t.Fatalf("generate token pair failed for user %d: %v", tt.userID, err)
		}

		claims, err := mgr.ValidateToken(accessToken)
		if err != nil {
			t.Fatalf("validate token failed for user %d: %v", tt.userID, err)
		}
		if claims.UserID != tt.userID {
			t.Errorf("expected user ID %d, got %d", tt.userID, claims.UserID)
		}
		if claims.AccountID != tt.accountID {
			t.Errorf("expected account ID %s, got %s", tt.accountID, claims.AccountID)
		}
	}
}
