package service

import (
	"testing"
	"time"
)

func TestGenerateTOTPSecret(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if secret == "" {
		t.Error("expected non-empty secret")
	}
	if len(secret) < 10 {
		t.Errorf("secret seems too short: %s", secret)
	}
}

func TestGenerateTOTPCode(t *testing.T) {
	secret, err := GenerateTOTPSecret()
	if err != nil {
		t.Fatalf("failed to generate secret: %v", err)
	}

	code, err := GenerateTOTPCode(secret, 0)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6-digit code, got %q (len=%d)", code, len(code))
	}
}

func TestGenerateTOTPCode_Deterministic(t *testing.T) {
	secret, _ := GenerateTOTPSecret()

	code1, _ := GenerateTOTPCode(secret, 12345)
	code2, _ := GenerateTOTPCode(secret, 12345)
	if code1 != code2 {
		t.Errorf("same timestamp should produce same code: %s vs %s", code1, code2)
	}
}

func TestGenerateTOTPCode_DifferentTimestamps(t *testing.T) {
	secret, _ := GenerateTOTPSecret()

	code1, _ := GenerateTOTPCode(secret, 0)
	code2, _ := GenerateTOTPCode(secret, 30)
	if code1 == code2 {
		t.Error("different time steps should likely produce different codes")
	}
}

func TestGenerateTOTPCode_InvalidSecret(t *testing.T) {
	_, err := GenerateTOTPCode("not-valid-base32!!!", 0)
	if err == nil {
		t.Error("expected error for invalid secret")
	}
}

func TestVerifyTOTPCode_Valid(t *testing.T) {
	secret, _ := GenerateTOTPSecret()

	code, err := GenerateTOTPCode(secret, time.Now().Unix())
	if err != nil {
		t.Fatalf("failed to generate code: %v", err)
	}

	valid := VerifyTOTPCode(secret, code)
	if !valid {
		t.Error("expected code to be valid")
	}
}

func TestVerifyTOTPCode_Invalid(t *testing.T) {
	secret, _ := GenerateTOTPSecret()

	valid := VerifyTOTPCode(secret, "000000")
	if valid {
		t.Error("expected random code to be invalid (unlikely to match)")
	}
}

func TestBuildOTPAuthURL(t *testing.T) {
	tests := []struct {
		name      string
		accountID string
		secret    string
		issuer    string
		want      string
	}{
		{
			name:      "standard url",
			accountID: "user123",
			secret:    "JBSWY3DPEHPK3PXP",
			issuer:    "TestApp",
			want:      "otpauth://totp/TestApp:user123?secret=JBSWY3DPEHPK3PXP&issuer=TestApp",
		},
		{
			name:      "different values",
			accountID: "alice",
			secret:    "ABCDEFGH",
			issuer:    "MyService",
			want:      "otpauth://totp/MyService:alice?secret=ABCDEFGH&issuer=MyService",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := BuildOTPAuthURL(tt.accountID, tt.secret, tt.issuer)
			if got != tt.want {
				t.Errorf("BuildOTPAuthURL() = %q, want %q", got, tt.want)
			}
		})
	}
}
