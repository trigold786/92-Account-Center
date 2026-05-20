package service

import (
	"context"
	"testing"
)

func TestGetAdConfigByLevel(t *testing.T) {
	svc := NewAdConfigService(nil)
	config, err := svc.GetAdConfig(context.Background(), "L3")
	if err != nil {
		t.Fatalf("GetAdConfig failed: %v", err)
	}
	if config.ShowAds {
		t.Fatal("L3 should not show ads")
	}
}

func TestAdConfigDefaults(t *testing.T) {
	svc := NewAdConfigService(nil)
	config, err := svc.GetAdConfig(context.Background(), "L0")
	if err != nil {
		t.Fatalf("GetAdConfig failed: %v", err)
	}
	if !config.ShowAds {
		t.Fatal("L0 should show ads")
	}
	if config.VideoMaxDurationSec != 5 {
		t.Fatalf("expected 5s video limit, got %d", config.VideoMaxDurationSec)
	}
}

func TestFrequencyControl(t *testing.T) {
	svc := NewAdConfigService(nil)
	allowed, err := svc.CheckFrequencyControl(context.Background(), "user_1", "splash")
	if err != nil {
		t.Fatalf("CheckFrequencyControl failed: %v", err)
	}
	if !allowed {
		t.Fatal("expected first request to be allowed")
	}
}
