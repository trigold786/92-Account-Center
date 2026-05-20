package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/repository"
)

func TestRegistrationTrends(t *testing.T) {
	repo := &mockMetricsRepo{}
	svc := NewMetricsService(repo)
	trends, err := svc.GetRegistrationTrends(context.Background(), "daily")
	if err != nil {
		t.Fatalf("GetRegistrationTrends failed: %v", err)
	}
	if len(trends) == 0 {
		t.Fatal("expected non-empty trends")
	}
	if trends[0].Count <= 0 {
		t.Fatal("expected positive count")
	}
}

func TestConversionFunnel(t *testing.T) {
	svc := NewMetricsService(nil)
	funnel, err := svc.GetConversionFunnel(context.Background())
	if err != nil {
		t.Fatalf("GetConversionFunnel failed: %v", err)
	}
	if len(funnel.Stages) == 0 {
		t.Fatal("expected non-empty funnel")
	}
	if funnel.Stages[0].Count < funnel.Stages[len(funnel.Stages)-1].Count {
		t.Fatal("funnel should narrow")
	}
}

func TestMRRCalculation(t *testing.T) {
	repo := &mockMetricsRepo{}
	svc := NewMetricsService(repo)
	mrr, err := svc.GetMRR(context.Background())
	if err != nil {
		t.Fatalf("GetMRR failed: %v", err)
	}
	if mrr.Total <= 0 {
		t.Fatal("expected positive MRR")
	}
}

type mockMetricsRepo struct{}

func (m *mockMetricsRepo) GetRegistrationCountByPeriod(ctx context.Context, period string) ([]repository.DateCount, error) {
	return []repository.DateCount{{Date: "2026-05-01", Count: 100}, {Date: "2026-05-02", Count: 120}}, nil
}

func (m *mockMetricsRepo) GetPaidUsers(ctx context.Context) (int, error) { return 50, nil }
func (m *mockMetricsRepo) GetMRR(ctx context.Context) (float64, error) { return 15000, nil }
func (m *mockMetricsRepo) GetTotalUsers(ctx context.Context) (int, error) { return 10000, nil }
func (m *mockMetricsRepo) GetSubscribedUsers(ctx context.Context) (int, error) { return 500, nil }
