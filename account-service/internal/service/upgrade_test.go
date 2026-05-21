package service

import (
	"context"
	"testing"
)

func TestUpgradePreview(t *testing.T) {
	svc := NewUpgradeService(nil, nil)
	preview, err := svc.PreviewUpgrade(context.Background(), 1, "basic", "pro")
	if err != nil {
		t.Fatalf("PreviewUpgrade failed: %v", err)
	}
	if preview.ImmediateTotal <= 0 {
		t.Fatal("expected positive upgrade fee")
	}
	if preview.ProratedDays <= 0 {
		t.Fatal("expected prorated days > 0")
	}
}

func TestDowngradePreview(t *testing.T) {
	svc := NewUpgradeService(nil, nil)
	preview, err := svc.PreviewDowngrade(context.Background(), 1, "pro", "basic")
	if err != nil {
		t.Fatalf("PreviewDowngrade failed: %v", err)
	}
	if preview.NextPeriodTotal <= 0 {
		t.Fatal("expected positive next period fee")
	}
	if !preview.EffectiveNextPeriod {
		t.Fatal("downgrade should be effective next period")
	}
}

func TestImmediateUpgrade(t *testing.T) {
	svc := NewUpgradeService(nil, nil)
	err := svc.ExecuteUpgrade(context.Background(), 1, "pro")
	if err != nil {
		t.Fatalf("ExecuteUpgrade failed: %v", err)
	}
}

type mockUpgradeSubRepo struct{}

func (m *mockUpgradeSubRepo) GetCurrentPlan(ctx context.Context, userID int64) (string, error) {
	return "basic", nil
}

func (m *mockUpgradeSubRepo) UpgradePlan(ctx context.Context, userID int64, newPlan string) error {
	return nil
}
