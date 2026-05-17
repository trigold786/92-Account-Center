package service

import "testing"

func TestNewRiskService(t *testing.T) {
	s := NewRiskService(nil, nil, nil)
	if s == nil {
		t.Error("NewRiskService returned nil")
	}
}

func TestNewAuditService(t *testing.T) {
	s := NewAuditService(nil, nil)
	if s == nil {
		t.Error("NewAuditService returned nil")
	}
}

func TestNewBlacklistService(t *testing.T) {
	s := NewBlacklistService(nil, nil, nil)
	if s == nil {
		t.Error("NewBlacklistService returned nil")
	}
}

func TestNewSlidingWindowLimiter(t *testing.T) {
	s := NewSlidingWindowLimiter(nil, nil)
	if s == nil {
		t.Error("NewSlidingWindowLimiter returned nil")
	}
}
