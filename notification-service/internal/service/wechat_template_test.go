package service

import (
	"context"
	"testing"
)

func TestValidateTemplateID_ValidType(t *testing.T) {
	svc := NewWeChatTemplateService("test_app", "test_secret", map[string]string{
		"subscription_expiring": "tpl_001",
	})
	err := svc.ValidateTemplateID("subscription_expiring")
	if err != nil {
		t.Fatalf("ValidateTemplateID failed: %v", err)
	}
}

func TestValidateTemplateID_InvalidType(t *testing.T) {
	svc := NewWeChatTemplateService("test_app", "test_secret", map[string]string{})
	err := svc.ValidateTemplateID("invalid_type")
	if err == nil {
		t.Fatal("expected error for invalid template type")
	}
}

func TestValidateTemplateID_MissingConfig(t *testing.T) {
	svc := NewWeChatTemplateService("test_app", "test_secret", map[string]string{})
	err := svc.ValidateTemplateID("payment_success")
	if err == nil {
		t.Fatal("expected error for unconfigured template")
	}
}

func TestSendSubscriptionMessage_NoCredentials(t *testing.T) {
	svc := NewWeChatTemplateService("", "", map[string]string{
		"subscription_expiring": "tpl_001",
	})
	err := svc.SendSubscriptionMessage(context.Background(), "user_openid", "subscription_expiring",
		map[string]string{"name": "专业版", "date": "2026-06-01"})
	if err == nil {
		t.Fatal("expected error when no credentials configured")
	}
}

func TestSendWithRetry_NoCredentials(t *testing.T) {
	svc := NewWeChatTemplateService("", "", map[string]string{
		"subscription_expiring": "tpl_001",
	})
	err := svc.SendWithRetry(context.Background(), "user_openid", "subscription_expiring",
		map[string]string{"name": "专业版"})
	if err == nil {
		t.Fatal("expected error when no credentials configured")
	}
}
