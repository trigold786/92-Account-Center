package service

import (
	"context"
	"fmt"
	"log"

	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
	"github.com/trigold786/92-Account-Center/credit-service/internal/repository"
)

type RebateService interface {
	ProcessSubscriptionPaid(ctx context.Context, event *model.ProcessSubscriptionPaidEvent) error
	GetRebateRate(ctx context.Context, subscriptionCount int) (float64, error)
}

type rebateService struct {
	creditRepo   repository.CreditRepository
	referralRepo repository.ReferralRepository
	creditSvc    CreditService
}

func NewRebateService(
	creditRepo repository.CreditRepository,
	referralRepo repository.ReferralRepository,
	creditSvc CreditService,
) RebateService {
	return &rebateService{
		creditRepo:   creditRepo,
		referralRepo: referralRepo,
		creditSvc:    creditSvc,
	}
}

func (s *rebateService) ProcessSubscriptionPaid(ctx context.Context, event *model.ProcessSubscriptionPaidEvent) error {
	ref, err := s.referralRepo.GetByRefereeID(ctx, event.RefereeID)
	if err != nil {
		return fmt.Errorf("lookup referral: %w", err)
	}
	if ref == nil {
		log.Printf("no referral relation for referee %d, skipping rebate", event.RefereeID)
		return nil
	}

	rate, err := s.GetRebateRate(ctx, ref.RefereeSubscriptionCount)
	if err != nil {
		return fmt.Errorf("get rebate rate: %w", err)
	}

	rewardAmount := event.SubscriptionPrice * rate
	if rewardAmount <= 0 {
		return nil
	}

	refID := fmt.Sprintf("rebate:%s:%d", event.OrderID, ref.ReferrerID)
	details := fmt.Sprintf(`{"type":"referral_rebate","referee_id":%d,"rate":%.4f,"order":"%s"}`, event.RefereeID, rate, event.OrderID)

	if err := s.creditSvc.EarnCredits(ctx, ref.ReferrerID, rewardAmount, "EARN_REFERRAL", refID, details); err != nil {
		return fmt.Errorf("earn rebate credits: %w", err)
	}

	if err := s.referralRepo.IncrementSubscriptionCount(ctx, event.RefereeID); err != nil {
		log.Printf("warning: failed to increment subscription count for referee %d: %v", event.RefereeID, err)
	}

	log.Printf("rebate processed: referrer=%d referee=%d rate=%.2f%% reward=%.2f", ref.ReferrerID, event.RefereeID, rate*100, rewardAmount)
	return nil
}

func (s *rebateService) GetRebateRate(ctx context.Context, subscriptionCount int) (float64, error) {
	cfg, err := s.creditRepo.GetRebateConfig(ctx, subscriptionCount)
	if err != nil {
		return 0, err
	}
	if cfg == nil {
		return 0.10, nil
	}
	return cfg.RebatePercentage, nil
}
