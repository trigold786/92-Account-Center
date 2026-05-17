package service

import "testing"

func TestNewCreditService(t *testing.T) {
	s := NewCreditService(nil, nil, nil)
	if s == nil {
		t.Error("NewCreditService returned nil")
	}
}

func TestNewRebateService(t *testing.T) {
	s := NewRebateService(nil, nil, nil, nil)
	if s == nil {
		t.Error("NewRebateService returned nil")
	}
}

func TestNewReferralService(t *testing.T) {
	s := NewReferralService(nil, nil)
	if s == nil {
		t.Error("NewReferralService returned nil")
	}
}
