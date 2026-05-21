package service

import (
	"context"
	"testing"
)

func TestSendSubscriptionMessage(t *testing.T) {
	svc := NewWeChatTemplateService(nil)
	err := svc.SendSubscriptionMessage(context.Background(), "user_openid", "subscription_expiring",
		map[string]string{"name": "专业版", "date": "2026-06-01"})
	if err != nil {
		t.Fatalf("SendSubscriptionMessage failed: %v", err)
	}
}

func TestValidateTemplateID(t *testing.T) {
	svc := NewWeChatTemplateService(nil)
	err := svc.ValidateTemplateID("subscription_expiring")
	if err != nil {
		t.Fatalf("ValidateTemplateID failed: %v", err)
	}
	err = svc.ValidateTemplateID("invalid_type")
	if err == nil {
		t.Fatal("expected error for invalid template type")
	}
}
