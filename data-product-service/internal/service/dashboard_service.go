package service

import (
	"context"
	"math"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/repository"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/svcconfig"
)

type DashboardService interface {
	GetOverview(ctx context.Context) (*model.DashboardOverview, error)
	GetSubscriptionFunnel(ctx context.Context) (*model.SubscriptionFunnel, error)
}

type dashboardService struct {
	dataRepo repository.DataRepository
	rfmSvc   RFMService
	cfg      *svcconfig.DataProductConfig
}

func NewDashboardService(dataRepo repository.DataRepository, rfmSvc RFMService, cfg *svcconfig.DataProductConfig) DashboardService {
	return &dashboardService{dataRepo: dataRepo, rfmSvc: rfmSvc, cfg: cfg}
}

func (s *dashboardService) GetOverview(ctx context.Context) (*model.DashboardOverview, error) {
	totalUsers, err := s.dataRepo.GetTotalUsers(ctx)
	if err != nil {
		return nil, err
	}

	totalSubs, err := s.dataRepo.GetTotalSubscriptions(ctx)
	if err != nil {
		return nil, err
	}

	earned, err := s.dataRepo.GetTotalCreditsByTypes(ctx, []string{"EARN_REFERRAL", "EARN_VERIFY", "REFUND_SUB"})
	if err != nil {
		return nil, err
	}

	consumed, err := s.dataRepo.GetTotalCreditsByTypes(ctx, []string{"CONSUME_SUB"})
	if err != nil {
		return nil, err
	}

	blCount, err := s.dataRepo.GetActiveBlacklistCount(ctx)
	if err != nil {
		return nil, err
	}

	trend, err := s.dataRepo.GetRegistrationTrend(ctx, s.cfg.DashboardTrendDays)
	if err != nil {
		return nil, err
	}

	flow, err := s.dataRepo.GetCreditFlow(ctx)
	if err != nil {
		return nil, err
	}

	rfmDist, err := s.rfmSvc.GetRFMDistribution(ctx)
	if err != nil {
		return nil, err
	}

	return &model.DashboardOverview{
		TotalUsers:           totalUsers,
		TotalSubscriptions:   totalSubs,
		TotalCreditsEarned:   earned,
		TotalCreditsConsumed: consumed,
		BlacklistActive:      blCount,
		RegistrationTrend:    trend,
		CreditFlow:           flow,
		RFMDistribution:      rfmDist,
	}, nil
}

func (s *dashboardService) GetSubscriptionFunnel(ctx context.Context) (*model.SubscriptionFunnel, error) {
	totalUsers, err := s.dataRepo.GetTotalUsers(ctx)
	if err != nil {
		return nil, err
	}
	if totalUsers == 0 {
		totalUsers = 1
	}

	tierCounts, err := s.dataRepo.GetUserTierCounts(ctx)
	if err != nil {
		return nil, err
	}

	tierMap := make(map[int]int)
	for _, tc := range tierCounts {
		tierMap[tc.Tier] = tc.Count
	}

	l1Plus := 0
	for tier, count := range tierMap {
		if tier >= 1 {
			l1Plus += count
		}
	}

	l2Plus, err := s.dataRepo.GetDistinctSubscriberCount(ctx, 2)
	if err != nil {
		return nil, err
	}

	l3Plus, err := s.dataRepo.GetDistinctSubscriberCount(ctx, 3)
	if err != nil {
		return nil, err
	}

	l4, err := s.dataRepo.GetDistinctSubscriberCount(ctx, 4)
	if err != nil {
		return nil, err
	}

	pct := func(n int) float64 {
		return math.Round(float64(n)*10000/float64(totalUsers)) / 100
	}

	return &model.SubscriptionFunnel{
		Steps: []model.FunnelStep{
			{Name: "注册用户", Count: totalUsers, Percentage: 100.0},
			{Name: "实名用户 (L1+)", Count: l1Plus, Percentage: pct(l1Plus)},
			{Name: "订阅用户 (L2+)", Count: l2Plus, Percentage: pct(l2Plus)},
			{Name: "高级订阅 (L3+)", Count: l3Plus, Percentage: pct(l3Plus)},
			{Name: "顶级订阅 (L4)", Count: l4, Percentage: pct(l4)},
		},
	}, nil
}
