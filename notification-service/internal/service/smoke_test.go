package service

import "testing"

func TestNewSMSService(t *testing.T) {
	s := NewSMSService(nil, nil, nil)
	if s == nil {
		t.Error("NewSMSService returned nil")
	}
}

func TestNewSimpleEmailService(t *testing.T) {
	s := NewSimpleEmailService(nil, nil)
	if s == nil {
		t.Error("NewSimpleEmailService returned nil")
	}
}

func TestNewOTPEmailService(t *testing.T) {
	s := NewOTPEmailService(nil, nil, "", "", nil)
	if s == nil {
		t.Error("NewOTPEmailService returned nil")
	}
}
