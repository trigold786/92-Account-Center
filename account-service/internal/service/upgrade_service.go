package service

import (
	"context"
	"math"
	"time"
)

type UpgradePreview struct {
	CurrentPlan     string  `json:"current_plan"`
	TargetPlan      string  `json:"target_plan"`
	ImmediateTotal  float64 `json:"immediate_total"`
	ProratedDays    int     `json:"prorated_days"`
	NextBillingDate string  `json:"next_billing_date"`
}

type DowngradePreview struct {
	CurrentPlan         string  `json:"current_plan"`
	TargetPlan          string  `json:"target_plan"`
	NextPeriodTotal     float64 `json:"next_period_total"`
	EffectiveNextPeriod bool    `json:"effective_next_period"`
}

var planPrices = map[string]float64{
	"basic":      9.9,
	"pro":        29.9,
	"enterprise": 99.9,
}

type UpgradeService struct {
	subRepo interface{}
	paySvc  interface{}
}

func NewUpgradeService(subRepo, paySvc interface{}) *UpgradeService {
	return &UpgradeService{subRepo: subRepo, paySvc: paySvc}
}

func (s *UpgradeService) PreviewUpgrade(ctx context.Context, userID int64, currentPlan, targetPlan string) (*UpgradePreview, error) {
	currentPrice := planPrices[currentPlan]
	targetPrice := planPrices[targetPlan]
	diff := targetPrice - currentPrice
	if diff <= 0 {
		diff = targetPrice
	}
	now := time.Now()
	daysInMonth := 30
	remainingDays := daysInMonth - now.Day()
	if remainingDays < 1 {
		remainingDays = 1
	}
	prorated := math.Round(diff*float64(remainingDays)/float64(daysInMonth)*100) / 100
	return &UpgradePreview{
		CurrentPlan:     currentPlan,
		TargetPlan:      targetPlan,
		ImmediateTotal:  prorated,
		ProratedDays:    remainingDays,
		NextBillingDate: now.AddDate(0, 1, 0).Format("2006-01-02"),
	}, nil
}

func (s *UpgradeService) PreviewDowngrade(ctx context.Context, userID int64, currentPlan, targetPlan string) (*DowngradePreview, error) {
	targetPrice := planPrices[targetPlan]
	return &DowngradePreview{
		CurrentPlan:         currentPlan,
		TargetPlan:          targetPlan,
		NextPeriodTotal:     targetPrice,
		EffectiveNextPeriod: true,
	}, nil
}

func (s *UpgradeService) ExecuteUpgrade(ctx context.Context, userID int64, targetPlan string) error {
	return nil
}
