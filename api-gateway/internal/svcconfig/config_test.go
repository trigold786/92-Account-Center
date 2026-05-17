package svcconfig

import "testing"

func TestGatewayConfigFields(t *testing.T) {
	cfg := &GatewayConfig{
		Port:                   "30300",
		RateLimitRPS:           100,
		CacheMaxAge:            300,
		MaxDesensitizeBodySize: 1048576,
	}
	if cfg.Port != "30300" {
		t.Errorf("expected Port 30300, got %s", cfg.Port)
	}
	if cfg.RateLimitRPS != 100 {
		t.Errorf("expected RateLimitRPS 100, got %d", cfg.RateLimitRPS)
	}
}

func TestGatewayConfigIsNotNil(t *testing.T) {
	cfg := &GatewayConfig{}
	if cfg == nil {
		t.Error("GatewayConfig should not be nil")
	}
}
