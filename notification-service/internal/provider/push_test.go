package provider

import (
	"context"
	"testing"
)

func TestPushProviderRegistry_RegisterAndGet(t *testing.T) {
	r := NewPushProviderRegistry()

	apns := NewAPNsProvider(APNsConfig{BundleID: "com.test.app"})
	fcm := NewFCMProvider(FCMConfig{ServerKey: "test-key", ProjectID: "test-project"})
	hms := NewHMSProvider(HMSConfig{AppID: "test-app-id", AppSecret: "test-secret"})

	r.Register(apns)
	r.Register(fcm)
	r.Register(hms)

	if p, ok := r.Get("apns"); !ok || p.Name() != "apns" {
		t.Error("expected to find apns provider")
	}
	if p, ok := r.Get("fcm"); !ok || p.Name() != "fcm" {
		t.Error("expected to find fcm provider")
	}
	if p, ok := r.Get("hms"); !ok || p.Name() != "hms" {
		t.Error("expected to find hms provider")
	}
	if _, ok := r.Get("nonexistent"); ok {
		t.Error("expected nonexistent provider to not be found")
	}
}

func TestPushProviderRegistry_List(t *testing.T) {
	r := NewPushProviderRegistry()

	if len(r.List()) != 0 {
		t.Error("expected empty registry")
	}

	r.Register(NewAPNsProvider(APNsConfig{}))
	r.Register(NewFCMProvider(FCMConfig{}))
	r.Register(NewHMSProvider(HMSConfig{}))

	names := r.List()
	if len(names) != 3 {
		t.Errorf("expected 3 providers, got %d", len(names))
	}

	found := map[string]bool{}
	for _, n := range names {
		found[n] = true
	}
	if !found["apns"] || !found["fcm"] || !found["hms"] {
		t.Errorf("expected all three providers, got %v", names)
	}
}

func TestAPNsProvider_Name(t *testing.T) {
	p := NewAPNsProvider(APNsConfig{BundleID: "com.test.app"})
	if p.Name() != "apns" {
		t.Errorf("expected name 'apns', got '%s'", p.Name())
	}
}

func TestAPNsProvider_Send(t *testing.T) {
	p := NewAPNsProvider(APNsConfig{BundleID: "com.test.app"})
	resp, err := p.Send(context.Background(), &PushRequest{
		DeviceToken: "test-token-123",
		Title:       "Test",
		Body:        "Test body",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
	if resp.MessageID == "" {
		t.Error("expected non-empty message ID")
	}
}

func TestAPNsProvider_ValidateToken(t *testing.T) {
	p := NewAPNsProvider(APNsConfig{})
	if err := p.ValidateToken(context.Background(), "valid-token"); err != nil {
		t.Errorf("expected valid token to pass, got: %v", err)
	}
	if err := p.ValidateToken(context.Background(), ""); err == nil {
		t.Error("expected empty token to fail validation")
	}
}

func TestFCMProvider_Name(t *testing.T) {
	p := NewFCMProvider(FCMConfig{ServerKey: "key", ProjectID: "proj"})
	if p.Name() != "fcm" {
		t.Errorf("expected name 'fcm', got '%s'", p.Name())
	}
}

func TestFCMProvider_Send(t *testing.T) {
	p := NewFCMProvider(FCMConfig{ServerKey: "key", ProjectID: "proj"})
	resp, err := p.Send(context.Background(), &PushRequest{
		DeviceToken: "fcm-token-456",
		Title:       "Hello",
		Body:        "World",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
	if resp.MessageID == "" {
		t.Error("expected non-empty message ID")
	}
}

func TestFCMProvider_ValidateToken(t *testing.T) {
	p := NewFCMProvider(FCMConfig{})
	if err := p.ValidateToken(context.Background(), "valid-token"); err != nil {
		t.Errorf("expected valid token to pass, got: %v", err)
	}
	if err := p.ValidateToken(context.Background(), ""); err == nil {
		t.Error("expected empty token to fail validation")
	}
}

func TestHMSProvider_Name(t *testing.T) {
	p := NewHMSProvider(HMSConfig{AppID: "app", AppSecret: "secret"})
	if p.Name() != "hms" {
		t.Errorf("expected name 'hms', got '%s'", p.Name())
	}
}

func TestHMSProvider_Send(t *testing.T) {
	p := NewHMSProvider(HMSConfig{AppID: "app", AppSecret: "secret"})
	resp, err := p.Send(context.Background(), &PushRequest{
		DeviceToken: "hms-token-789",
		Title:       "Push",
		Body:        "Notification",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Error("expected success")
	}
	if resp.MessageID == "" {
		t.Error("expected non-empty message ID")
	}
}

func TestHMSProvider_ValidateToken(t *testing.T) {
	p := NewHMSProvider(HMSConfig{})
	if err := p.ValidateToken(context.Background(), "valid-token"); err != nil {
		t.Errorf("expected valid token to pass, got: %v", err)
	}
	if err := p.ValidateToken(context.Background(), ""); err == nil {
		t.Error("expected empty token to fail validation")
	}
}

func TestPushRequest_Fields(t *testing.T) {
	badge := 3
	req := &PushRequest{
		DeviceToken: "token",
		Title:       "title",
		Body:        "body",
		Data:        map[string]string{"key": "value"},
		Priority:    "high",
		Sound:       "default",
		Badge:       &badge,
	}
	if req.DeviceToken != "token" {
		t.Error("DeviceToken mismatch")
	}
	if req.Priority != "high" {
		t.Error("Priority mismatch")
	}
	if *req.Badge != 3 {
		t.Error("Badge mismatch")
	}
}
