package service

import "testing"

func TestNewDashboardService(t *testing.T) {
	s := NewDashboardService(nil, nil, nil)
	if s == nil {
		t.Error("NewDashboardService returned nil")
	}
}
