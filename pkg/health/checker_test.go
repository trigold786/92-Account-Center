package health

import (
	"context"
	"errors"
	"testing"
)

func TestComponentHealthStatus(t *testing.T) {
	tests := []struct {
		name   string
		status Status
		wantOK bool
	}{
		{"up", StatusUp, true},
		{"degraded", StatusDegraded, true},
		{"down", StatusDown, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := ComponentHealth{Name: "test", Status: tt.status}
			if (ch.Status == StatusUp || ch.Status == StatusDegraded) != tt.wantOK {
				t.Errorf("unexpected ok state for status %v", tt.status)
			}
		})
	}
}

func TestCompositeChecker(t *testing.T) {
	ok := &mockChecker{name: "ok", status: StatusUp}
	fail := &mockChecker{name: "fail", status: StatusDown, err: errors.New("fail")}

	allOK := CompositeChecker{Checkers: []Checker{ok}}
	result := allOK.Check(context.Background())
	if result.Status != StatusUp {
		t.Fatalf("expected up, got %v", result.Status)
	}

	withFail := CompositeChecker{Checkers: []Checker{ok, fail}}
	result2 := withFail.Check(context.Background())
	if result2.Status != StatusDown {
		t.Fatalf("expected down, got %v", result2.Status)
	}
	if len(result2.Checks) != 2 {
		t.Fatalf("expected 2 checks, got %d", len(result2.Checks))
	}
}

type mockChecker struct {
	name   string
	status Status
	err    error
}

func (m *mockChecker) Name() string { return m.name }

func (m *mockChecker) Check(ctx context.Context) ComponentHealth {
	ch := ComponentHealth{Name: m.name, Status: m.status}
	if m.err != nil {
		ch.Error = m.err.Error()
	}
	return ch
}
