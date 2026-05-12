package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/trigold786/92-Account-Center/sms-email-service/internal/provider"
	"github.com/trigold786/92-Account-Center/sms-email-service/pkg/circuitbreaker"
)

type mockSMSProvider struct {
	name string
	code string
	err  error
}

func (m *mockSMSProvider) SendCode(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.code, nil
}

func (m *mockSMSProvider) Name() string {
	return m.name
}

func TestGetProviderStatus(t *testing.T) {
	p1 := &mockSMSProvider{name: "mock1", code: "123456", err: nil}
	p2 := &mockSMSProvider{name: "mock2", code: "654321", err: nil}
	cb1 := circuitbreaker.New(5, 30*time.Second)
	cb2 := circuitbreaker.New(5, 30*time.Second)

	svc := NewSMSService([]provider.SMSProvider{p1, p2}, []*circuitbreaker.CircuitBreaker{cb1, cb2})

	statuses := svc.GetProviderStatus()
	if len(statuses) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(statuses))
	}
	if statuses[0].Name != "mock1" {
		t.Errorf("expected mock1, got %s", statuses[0].Name)
	}
	if statuses[0].Status != "closed" {
		t.Errorf("expected closed status, got %s", statuses[0].Status)
	}
	if statuses[1].Name != "mock2" {
		t.Errorf("expected mock2, got %s", statuses[1].Name)
	}
}

func TestGetProviderStatus_OpenCircuit(t *testing.T) {
	p1 := &mockSMSProvider{name: "failing", err: errors.New("fail")}
	cb1 := circuitbreaker.New(1, 30*time.Second)

	svc := NewSMSService([]provider.SMSProvider{p1}, []*circuitbreaker.CircuitBreaker{cb1})

	for i := 0; i < 5; i++ {
		cb1.Execute(func() error { return errors.New("fail") })
	}

	statuses := svc.GetProviderStatus()
	if len(statuses) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(statuses))
	}
	if statuses[0].Status != "open" {
		t.Errorf("expected open status, got %s", statuses[0].Status)
	}
}

func TestMockProvider_Name(t *testing.T) {
	p := &mockSMSProvider{name: "test-provider"}
	if p.Name() != "test-provider" {
		t.Errorf("expected test-provider, got %s", p.Name())
	}
}

func TestMockProvider_SendCode(t *testing.T) {
	p := &mockSMSProvider{name: "test", code: "123456", err: nil}
	code, err := p.SendCode(context.Background(), "1234567890")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if code != "123456" {
		t.Errorf("expected 123456, got %s", code)
	}

	p2 := &mockSMSProvider{name: "fail", err: errors.New("network error")}
	_, err = p2.SendCode(context.Background(), "1234567890")
	if err == nil {
		t.Error("expected error from failing provider")
	}
}

func TestGenerateEmailCode_Length(t *testing.T) {
	for _, length := range []int{4, 6, 8} {
		code := generateEmailCode(length)
		if len(code) != length {
			t.Errorf("expected length %d, got %d", length, len(code))
		}
	}
}

func TestGenerateEmailCode_AllDigits(t *testing.T) {
	for i := 0; i < 100; i++ {
		code := generateEmailCode(6)
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Errorf("expected all digits, got %c in %s", c, code)
			}
		}
	}
}
