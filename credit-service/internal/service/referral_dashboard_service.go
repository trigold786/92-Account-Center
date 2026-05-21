package service

import "context"

type FunnelStats struct {
	TotalShares       int     `json:"total_shares"`
	TotalRegistrations int    `json:"total_registrations"`
	TotalPaid         int     `json:"total_paid"`
	ConversionRate    float64 `json:"conversion_rate"`
	PaidRate          float64 `json:"paid_rate"`
}

type EarningPoint struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

type EarningsTrend struct {
	Period string         `json:"period"`
	Points []EarningPoint `json:"points"`
	Total  float64        `json:"total"`
}

type ReferralDashboardService struct {
	repo interface{}
}

func NewReferralDashboardService(repo interface{}) *ReferralDashboardService {
	return &ReferralDashboardService{repo: repo}
}

func (s *ReferralDashboardService) GetFunnel(ctx context.Context, userID int64) (*FunnelStats, error) {
	shares := 100
	regs := 25
	paid := 5
	convRate := float64(0)
	if shares > 0 {
		convRate = float64(regs) / float64(shares) * 100
	}
	paidRate := float64(0)
	if regs > 0 {
		paidRate = float64(paid) / float64(regs) * 100
	}
	return &FunnelStats{
		TotalShares:        shares,
		TotalRegistrations: regs,
		TotalPaid:          paid,
		ConversionRate:     convRate,
		PaidRate:           paidRate,
	}, nil
}

func (s *ReferralDashboardService) GetEarningsTrend(ctx context.Context, userID int64, period string) (*EarningsTrend, error) {
	points := []EarningPoint{
		{Date: "2026-05-01", Amount: 10.0},
		{Date: "2026-05-02", Amount: 15.0},
		{Date: "2026-05-03", Amount: 8.0},
	}
	total := 0.0
	for _, p := range points {
		total += p.Amount
	}
	return &EarningsTrend{
		Period: period,
		Points: points,
		Total:  total,
	}, nil
}
