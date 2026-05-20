package auth

import (
	"testing"
)

func TestHashPasswordArgon2id_VerifySuccess(t *testing.T) {
	hash, err := HashPasswordArgon2id("testpassword123", DefaultParams())
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}
	if !IsArgon2idHash(hash) {
		t.Error("expected argon2id prefix")
	}

	match, err := VerifyPasswordArgon2id("testpassword123", hash)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if !match {
		t.Error("expected password to match")
	}
}

func TestHashPasswordArgon2id_WrongPassword(t *testing.T) {
	hash, err := HashPasswordArgon2id("correctpass", DefaultParams())
	if err != nil {
		t.Fatalf("hash failed: %v", err)
	}

	match, err := VerifyPasswordArgon2id("wrongpass", hash)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if match {
		t.Error("expected password mismatch")
	}
}

func TestHashPasswordArgon2id_UniqueSalts(t *testing.T) {
	hash1, _ := HashPasswordArgon2id("samepass", DefaultParams())
	hash2, _ := HashPasswordArgon2id("samepass", DefaultParams())
	if hash1 == hash2 {
		t.Error("hashes should differ due to random salts")
	}
}

func TestIsArgon2idHash(t *testing.T) {
	tests := []struct {
		hash string
		want bool
	}{
		{"$argon2id$v=19$m=65536,t=3,p=2$salt$hash", true},
		{"testsalt$" + "ab12cd34", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsArgon2idHash(tt.hash); got != tt.want {
			t.Errorf("IsArgon2idHash(%q) = %v, want %v", tt.hash[:20], got, tt.want)
		}
	}
}

func TestVerifyPassword_SM3Compat(t *testing.T) {
	salt := "testsalt"
	sm3Hash := salt + "$" + "fakesm3hashvalue"

	if IsArgon2idHash(sm3Hash) {
		t.Error("SM3 hash should not be identified as argon2id")
	}
	if IdentifyHashAlgorithm(sm3Hash) != HashSM3 {
		t.Error("expected SM3 algorithm")
	}
	if IdentifyHashAlgorithm("$argon2id$v=19$m=65536,t=3,p=2$salt$hash") != HashArgon2id {
		t.Error("expected argon2id algorithm")
	}
}

func TestHashPassword_Integration(t *testing.T) {
	result, err := HashPassword("mypassword")
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}
	if result.Algorithm != HashArgon2id {
		t.Errorf("expected argon2id, got %s", result.Algorithm)
	}

	if !VerifyPassword("mypassword", result.Hash) {
		t.Error("VerifyPassword should succeed for correct password")
	}
	if VerifyPassword("wrongpassword", result.Hash) {
		t.Error("VerifyPassword should fail for wrong password")
	}
}

func TestNeedsRehash(t *testing.T) {
	if NeedsRehash("$argon2id$v=19$m=65536,t=3,p=2$s$hash") {
		t.Error("argon2id hash should not need rehash")
	}
	if !NeedsRehash("salt$sm3hash") {
		t.Error("SM3 hash should need rehash")
	}
}
