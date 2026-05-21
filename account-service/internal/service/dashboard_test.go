package service

import (
	"context"
	"testing"
)

func TestDashboardByLevel(t *testing.T) {
	svc := NewDashboardService(nil)

	dash, err := svc.GetDashboard(context.Background(), 0, "L0")
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}
	if len(dash.Cards) == 0 {
		t.Fatal("expected at least 1 card for L0")
	}
	found := false
	for _, c := range dash.Cards {
		if c.Type == "upgrade_guide" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("L0 should have upgrade_guide card")
	}

	dash2, err := svc.GetDashboard(context.Background(), 1, "L2")
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}
	foundCredit := false
	for _, c := range dash2.Cards {
		if c.Type == "credit_balance" {
			foundCredit = true
			break
		}
	}
	if !foundCredit {
		t.Fatal("L2 should have credit_balance card")
	}
}

func TestDashboardConfigFromService(t *testing.T) {
	svc := NewDashboardService(nil)

	dash, err := svc.GetDashboard(context.Background(), 2, "L4")
	if err != nil {
		t.Fatalf("GetDashboard failed: %v", err)
	}
	exclusiveCards := []string{"admin_panel", "enterprise_settings"}
	cardTypes := make(map[string]bool)
	for _, c := range dash.Cards {
		cardTypes[c.Type] = true
	}
	for _, card := range exclusiveCards {
		if !cardTypes[card] {
			t.Fatalf("L4 missing card: %s", card)
		}
	}
}
