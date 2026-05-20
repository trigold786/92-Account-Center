package vault

import (
	"context"
	"testing"
)

func TestInMemoryVault(t *testing.T) {
	v := NewInMemoryVault()
	err := v.SetSecret(context.Background(), "jwt/signing-key", "my-secret-key-123")
	if err != nil {
		t.Fatalf("SetSecret failed: %v", err)
	}
	val, err := v.GetSecret(context.Background(), "jwt/signing-key")
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}
	if val != "my-secret-key-123" {
		t.Fatalf("unexpected value: %s", val)
	}
}

func TestVaultRotate(t *testing.T) {
	v := NewInMemoryVault()
	v.SetSecret(context.Background(), "db/password", "old-pass")
	v.RotateSecret(context.Background(), "db/password")
	newPass, _ := v.GetSecret(context.Background(), "db/password")
	if newPass == "old-pass" {
		t.Fatal("password should have changed after rotation")
	}
}

func TestKeyExpiry(t *testing.T) {
	v := NewInMemoryVault()
	v.SetSecret(context.Background(), "temp/key", "temp-value")
	v.SetExpiry(context.Background(), "temp/key")
	val, err := v.GetSecret(context.Background(), "temp/key")
	if err == nil {
		t.Fatalf("expected error for expired key, got value: %s", val)
	}
}
