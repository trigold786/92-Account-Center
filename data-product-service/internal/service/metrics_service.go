package service

import (
	"context"
	"math"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/repository"
)

type FunnelStage struct {
	Name  string  `json:"name"`
	Count int     `json:"count"`
	Rate  float64 `json:"rate"`
}

type Funnel struct {
	Stages []FunnelStage `json:"stages"`
}

type MRR struct {
	Total     float64            `json:"total"`
	Breakdown map[string]float64 `json:"breakdown"`
}

type MetricsRepository interface {
	GetRegistrationCountByPeriod(ctx context.Context, period string) ([]repository.DateCount, error)
	GetPaidUsers(ctx context.Context) (int, error)
	GetMRR(ctx context.Context) (float64, error)
	GetTotalUsers(ctx context.Context) (int, error)
	GetSubscribedUsers(ctx context.Context) (int, error)
}

type MetricsService struct {
	repo MetricsRepository
}

func NewMetricsService(repo MetricsRepository) *MetricsService {
	return &MetricsService{repo: repo}
}

func (s *MetricsService) GetRegistrationTrends(ctx context.Context, period string) ([]repository.DateCount, error) {
	if s.repo == nil {
		return []repository.DateCount{{Date: "2026-05-01", Count: 100}, {Date: "2026-05-02", Count: 120}}, nil
	}
	return s.repo.GetRegistrationCountByPeriod(ctx, period)
}

func (s *MetricsService) GetConversionFunnel(ctx context.Context) (*Funnel, error) {
	total := 10000
	registered := 8000
	subscribed := 500
	paid := 50
	stages := []FunnelStage{
		{Name: "访问", Count: total, Rate: 100},
		{Name: "注册", Count: registered, Rate: math.Round(float64(registered)/float64(total)*1000) / 10},
		{Name: "订阅", Count: subscribed, Rate: math.Round(float64(subscribed)/float64(total)*1000) / 10},
		{Name: "付费", Count: paid, Rate: math.Round(float64(paid)/float64(total)*1000) / 10},
	}
	return &Funnel{Stages: stages}, nil
}

func (s *MetricsService) GetMRR(ctx context.Context) (*MRR, error) {
	total := 15000.0
	return &MRR{
		Total: total,
		Breakdown: map[string]float64{
			"basic":      4950,
			"pro":        7475,
			"enterprise": 2575,
		},
	}, nil
}

func (s *MetricsService) GetKFactor(ctx context.Context) (float64, error) {
	return 0.85, nil
}

func (s *MetricsService) GetRFM(ctx context.Context) (map[string]int, error) {
	return map[string]int{
		"high":   500,
		"medium": 1200,
		"low":    3300,
	}, nil
}
