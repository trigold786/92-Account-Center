package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
)

func TestCreateStrategy(t *testing.T) {
	svc := NewPushStrategyService(nil)

	st, err := svc.CreateStrategy(context.Background(), &model.PushStrategy{
		Name:         "welcome_push",
		FrequencyCap: 3,
		DNDStartHour: 22,
		DNDEndHour:   8,
		Enabled:      true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if st.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestFrequencyCap(t *testing.T) {
	svc := NewPushStrategyService(nil)

	st, _ := svc.CreateStrategy(context.Background(), &model.PushStrategy{
		Name:         "capped",
		FrequencyCap: 2,
		DNDStartHour: 0,
		DNDEndHour:   0,
		Enabled:      true,
	})

	allowed, _ := svc.EvaluateStrategy(context.Background(), st.ID, 1)
	if !allowed {
		t.Fatal("expected first push to be allowed")
	}

	allowed, _ = svc.EvaluateStrategy(context.Background(), st.ID, 1)
	if !allowed {
		t.Fatal("expected second push to be allowed")
	}

	allowed, reason := svc.EvaluateStrategy(context.Background(), st.ID, 1)
	if allowed {
		t.Fatal("expected third push to be blocked")
	}
	if reason != "frequency cap exceeded" {
		t.Fatalf("expected frequency cap exceeded, got %s", reason)
	}
}

func TestDNDSkip(t *testing.T) {
	svc := NewPushStrategyService(nil)

	st, _ := svc.CreateStrategy(context.Background(), &model.PushStrategy{
		Name:         "dnd_test",
		FrequencyCap: 0,
		DNDStartHour: 0,
		DNDEndHour:   24,
		Enabled:      true,
	})

	allowed, reason := svc.EvaluateStrategy(context.Background(), st.ID, 1)
	if allowed {
		t.Fatal("expected DND block")
	}
	if reason != "DND hours" {
		t.Fatalf("expected DND hours, got %s", reason)
	}
}

func TestDisabledStrategy(t *testing.T) {
	svc := NewPushStrategyService(nil)

	st, _ := svc.CreateStrategy(context.Background(), &model.PushStrategy{
		Name:    "disabled",
		Enabled: false,
	})

	allowed, reason := svc.EvaluateStrategy(context.Background(), st.ID, 1)
	if allowed {
		t.Fatal("expected disabled block")
	}
	if reason != "strategy disabled" {
		t.Fatalf("expected strategy disabled, got %s", reason)
	}
}

func TestNonExistentStrategy(t *testing.T) {
	svc := NewPushStrategyService(nil)

	allowed, reason := svc.EvaluateStrategy(context.Background(), 999, 1)
	if allowed {
		t.Fatal("expected not found block")
	}
	if reason != "strategy not found" {
		t.Fatalf("expected strategy not found, got %s", reason)
	}
}
