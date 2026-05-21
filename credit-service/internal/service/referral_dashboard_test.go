package service

import (
	"context"
	"testing"
)

func TestReferralFunnel(t *testing.T) {
	svc := NewReferralDashboardService(nil)
	userID := int64(1)
	funnel, err := svc.GetFunnel(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetFunnel failed: %v", err)
	}
	if funnel.TotalShares == 0 {
		t.Fatal("expected non-zero total shares")
	}
	if funnel.TotalShares < funnel.TotalRegistrations {
		t.Fatal("shares should be >= registrations")
	}
	if funnel.TotalRegistrations < funnel.TotalPaid {
		t.Fatal("registrations should be >= paid conversions")
	}
}

func TestEarningsTrend(t *testing.T) {
	svc := NewReferralDashboardService(nil)
	userID := int64(1)
	trend, err := svc.GetEarningsTrend(context.Background(), userID, "weekly")
	if err != nil {
		t.Fatalf("GetEarningsTrend failed: %v", err)
	}
	if len(trend.Points) == 0 {
		t.Fatal("expected non-empty trend points")
	}
}

type mockDashReferralRepo struct{}

func (m *mockDashReferralRepo) GetFunnelStats(ctx context.Context, userID int64) (shares, regs, paid int, err error) {
	return 100, 25, 5, nil
}

func (m *mockDashReferralRepo) GetEarningsHistory(ctx context.Context, userID int64, period string) ([]EarningPoint, error) {
	return []EarningPoint{
		{Date: "2026-05-01", Amount: 10.0},
		{Date: "2026-05-02", Amount: 15.0},
	}, nil
}
