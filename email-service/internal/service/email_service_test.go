package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/trigold786/92-Account-Center/email-service/internal/provider"
)

type mockEmailProvider struct {
	name    string
	result  *provider.EmailResult
}

func (m *mockEmailProvider) Send(_ context.Context, _, _, _ string) *provider.EmailResult {
	return m.result
}

func (m *mockEmailProvider) Name() string {
	return m.name
}

func TestGenerateOTP_Length(t *testing.T) {
	otp := generateOTP()
	if len(otp) != 6 {
		t.Errorf("expected 6-digit OTP, got length %d", len(otp))
	}
}

func TestGenerateOTP_AllDigits(t *testing.T) {
	for i := 0; i < 100; i++ {
		otp := generateOTP()
		for _, c := range otp {
			if c < '0' || c > '9' {
				t.Errorf("expected all digits, got %c in %s", c, otp)
			}
		}
	}
}

func TestMockEmailProvider_Success(t *testing.T) {
	p := &mockEmailProvider{
		name: "test",
		result: &provider.EmailResult{Success: true, MessageID: "msg-123"},
	}
	result := p.Send(context.Background(), "test@example.com", "Subject", "Body")
	if !result.Success {
		t.Error("expected success")
	}
	if result.MessageID != "msg-123" {
		t.Errorf("expected msg-123, got %s", result.MessageID)
	}
}

func TestMockEmailProvider_Failure(t *testing.T) {
	p := &mockEmailProvider{
		name: "fail",
		result: &provider.EmailResult{Success: false, Error: errors.New("smtp error")},
	}
	result := p.Send(context.Background(), "test@example.com", "Subject", "Body")
	if result.Success {
		t.Error("expected failure")
	}
	if result.Error == nil {
		t.Error("expected error")
	}
}

func TestMockEmailProvider_Name(t *testing.T) {
	p := &mockEmailProvider{name: "sendgrid"}
	if p.Name() != "sendgrid" {
		t.Errorf("expected sendgrid, got %s", p.Name())
	}
}

func TestSendEmail_Success(t *testing.T) {
	p := &mockEmailProvider{
		name: "test",
		result: &provider.EmailResult{Success: true},
	}
	svc := NewEmailService(nil, p, "secret", "from@test.com")

	err := svc.SendEmail(context.Background(), "to@test.com", "Test Subject", "Test Content")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSendEmail_Failure(t *testing.T) {
	p := &mockEmailProvider{
		name: "fail",
		result: &provider.EmailResult{Success: false, Error: errors.New("smtp error")},
	}
	svc := NewEmailService(nil, p, "secret", "from@test.com")

	err := svc.SendEmail(context.Background(), "to@test.com", "Subject", "Content")
	if err == nil {
		t.Error("expected error for failed send")
	}
	if !strings.Contains(err.Error(), "smtp error") {
		t.Errorf("expected smtp error message, got %v", err)
	}
}

func TestVerifyMagicLink_InvalidToken(t *testing.T) {
	p := &mockEmailProvider{
		name: "test",
		result: &provider.EmailResult{Success: true},
	}
	svc := NewEmailService(nil, p, "test-secret-key-123", "from@test.com")

	_, err := svc.VerifyMagicLink(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestSendMagicLink_Success(t *testing.T) {
	p := &mockEmailProvider{
		name: "test",
		result: &provider.EmailResult{Success: true},
	}
	svc := NewEmailService(nil, p, "test-secret-key-12345", "from@test.com")

	resp, err := svc.SendMagicLink(context.Background(), "user@test.com", "https://example.com/login")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp.MagicLink == "" {
		t.Error("expected magic link URL")
	}
	if !strings.HasPrefix(resp.MagicLink, "https://example.com/login?token=") {
		t.Errorf("magic link should start with target URL, got %s", resp.MagicLink)
	}
	if resp.ExpiresIn != 900 {
		t.Errorf("expected 900 seconds, got %d", resp.ExpiresIn)
	}
}

func TestMagicLink_Roundtrip(t *testing.T) {
	p := &mockEmailProvider{
		name: "test",
		result: &provider.EmailResult{Success: true},
	}
	secret := "test-roundtrip-secret-"
	svc := NewEmailService(nil, p, secret, "from@test.com")

	resp, err := svc.SendMagicLink(context.Background(), "roundtrip@test.com", "https://example.com/auth")
	if err != nil {
		t.Fatalf("SendMagicLink failed: %v", err)
	}

	parts := strings.SplitN(resp.MagicLink, "token=", 2)
	if len(parts) != 2 {
		t.Fatal("magic link should contain token parameter")
	}
	token := parts[1]

	email, err := svc.VerifyMagicLink(context.Background(), token)
	if err != nil {
		t.Fatalf("VerifyMagicLink failed: %v", err)
	}
	if email != "roundtrip@test.com" {
		t.Errorf("expected roundtrip@test.com, got %s", email)
	}
}
