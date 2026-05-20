package service

import (
	"context"
	"testing"
	"time"
)

func TestBiometricTokenBinding(t *testing.T) {
	svc := NewBiometricService(&mockBiometricRepo{})
	userID := int64(42)
	deviceID := "device_abc"
	token := "bio_token_123"

	err := svc.BindDeviceToken(context.Background(), userID, deviceID, token)
	if err != nil {
		t.Fatalf("BindDeviceToken failed: %v", err)
	}
}

func TestBiometricTokenRotation(t *testing.T) {
	svc := NewBiometricService(&mockBiometricRepo{})
	userID := int64(42)
	oldToken := "old_token"
	newToken, err := svc.RotateToken(context.Background(), userID, oldToken)
	if err != nil {
		t.Fatalf("RotateToken failed: %v", err)
	}
	if newToken == oldToken {
		t.Fatal("new token should differ from old token")
	}
}

func TestExpiredTokenFallsBack(t *testing.T) {
	svc := NewBiometricService(&mockBiometricRepo{
		tokenExpired: true,
	})
	userID := int64(42)
	deviceID := "device_abc"
	token := "expired_token"
	valid, err := svc.ValidateToken(context.Background(), userID, deviceID, token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if valid {
		t.Fatal("expected expired token to be invalid")
	}
}

type mockBiometricRepo struct {
	tokenExpired bool
}

func (m *mockBiometricRepo) GetDeviceToken(ctx context.Context, userID int64, deviceID string) (string, error) {
	if m.tokenExpired {
		return "", nil
	}
	return "stored_token", nil
}

func (m *mockBiometricRepo) SaveDeviceToken(ctx context.Context, userID int64, deviceID, token string, expiry time.Duration) error {
	return nil
}

func (m *mockBiometricRepo) DeleteDeviceToken(ctx context.Context, userID int64, deviceID string) error {
	return nil
}
