package service

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
	"github.com/trigold786/92-Account-Center/credit-service/internal/svcconfig"
)

func TestRebateService_ProcessSubscriptionPaid_NoReferral(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	referralRepo.On("GetByRefereeID", mock.Anything, int64(5)).Return(nil, nil)

	err := svc.ProcessSubscriptionPaid(context.Background(), &model.ProcessSubscriptionPaidEvent{
		RefereeID:         5,
		SubscriptionPrice: 100.0,
		OrderID:           "ORD-001",
	})
	assert.NoError(t, err)
	referralRepo.AssertExpectations(t)
	creditSvc.AssertNotCalled(t, "EarnCredits")
}

func TestRebateService_ProcessSubscriptionPaid_Success(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	ref := &model.ReferralRelation{
		ID:                       1,
		ReferrerID:               10,
		RefereeID:                5,
		RefereeSubscriptionCount: 3,
	}
	referralRepo.On("GetByRefereeID", mock.Anything, int64(5)).Return(ref, nil)

	creditRepo.On("GetRebateConfig", mock.Anything, 3).
		Return(&model.RebateConfig{RebatePercentage: 0.15}, nil)

	creditSvc.On("EarnCredits",
		mock.Anything,
		int64(10),
		15.0,
		"EARN_REFERRAL",
		"rebate:ORD-001:10",
		mock.MatchedBy(func(d string) bool {
			return len(d) > 0
		}),
	).Return(nil)

	referralRepo.On("IncrementSubscriptionCount", mock.Anything, int64(5)).Return(nil)

	err := svc.ProcessSubscriptionPaid(context.Background(), &model.ProcessSubscriptionPaidEvent{
		RefereeID:         5,
		SubscriptionPrice: 100.0,
		OrderID:           "ORD-001",
	})
	assert.NoError(t, err)
	creditRepo.AssertExpectations(t)
	referralRepo.AssertExpectations(t)
	creditSvc.AssertExpectations(t)
}

func TestRebateService_ProcessSubscriptionPaid_ZeroReward(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	ref := &model.ReferralRelation{
		ID:                       1,
		ReferrerID:               10,
		RefereeID:                5,
		RefereeSubscriptionCount: 3,
	}
	referralRepo.On("GetByRefereeID", mock.Anything, int64(5)).Return(ref, nil)

	creditRepo.On("GetRebateConfig", mock.Anything, 3).
		Return(&model.RebateConfig{RebatePercentage: 0.0}, nil)

	err := svc.ProcessSubscriptionPaid(context.Background(), &model.ProcessSubscriptionPaidEvent{
		RefereeID:         5,
		SubscriptionPrice: 100.0,
		OrderID:           "ORD-002",
	})
	assert.NoError(t, err)
	creditSvc.AssertNotCalled(t, "EarnCredits")
}

func TestRebateService_ProcessSubscriptionPaid_GetReferralError(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	referralRepo.On("GetByRefereeID", mock.Anything, int64(5)).
		Return(nil, errors.New("db error"))

	err := svc.ProcessSubscriptionPaid(context.Background(), &model.ProcessSubscriptionPaidEvent{
		RefereeID:         5,
		SubscriptionPrice: 100.0,
		OrderID:           "ORD-003",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "lookup referral")
	referralRepo.AssertExpectations(t)
}

func TestRebateService_ProcessSubscriptionPaid_GetRebateRateError(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	ref := &model.ReferralRelation{
		ID:                       1,
		ReferrerID:               10,
		RefereeID:                5,
		RefereeSubscriptionCount: 3,
	}
	referralRepo.On("GetByRefereeID", mock.Anything, int64(5)).Return(ref, nil)
	creditRepo.On("GetRebateConfig", mock.Anything, 3).
		Return(nil, errors.New("config error"))

	err := svc.ProcessSubscriptionPaid(context.Background(), &model.ProcessSubscriptionPaidEvent{
		RefereeID:         5,
		SubscriptionPrice: 100.0,
		OrderID:           "ORD-004",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "get rebate rate")
}

func TestRebateService_ProcessSubscriptionPaid_EarnCreditsError(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	ref := &model.ReferralRelation{
		ID:                       1,
		ReferrerID:               10,
		RefereeID:                5,
		RefereeSubscriptionCount: 1,
	}
	referralRepo.On("GetByRefereeID", mock.Anything, int64(5)).Return(ref, nil)
	creditRepo.On("GetRebateConfig", mock.Anything, 1).
		Return(&model.RebateConfig{RebatePercentage: 0.10}, nil)
	creditSvc.On("EarnCredits",
		mock.Anything, int64(10), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(errors.New("tx failed"))

	err := svc.ProcessSubscriptionPaid(context.Background(), &model.ProcessSubscriptionPaidEvent{
		RefereeID:         5,
		SubscriptionPrice: 100.0,
		OrderID:           "ORD-005",
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "earn rebate credits")
}

func TestRebateService_ProcessSubscriptionPaid_IncrementSubscriptionCountWarning(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	ref := &model.ReferralRelation{
		ID:                       1,
		ReferrerID:               10,
		RefereeID:                5,
		RefereeSubscriptionCount: 1,
	}
	referralRepo.On("GetByRefereeID", mock.Anything, int64(5)).Return(ref, nil)
	creditRepo.On("GetRebateConfig", mock.Anything, 1).
		Return(&model.RebateConfig{RebatePercentage: 0.10}, nil)
	creditSvc.On("EarnCredits",
		mock.Anything, int64(10), mock.Anything, mock.Anything, mock.Anything, mock.Anything,
	).Return(nil)
	referralRepo.On("IncrementSubscriptionCount", mock.Anything, int64(5)).
		Return(errors.New("update failed"))

	err := svc.ProcessSubscriptionPaid(context.Background(), &model.ProcessSubscriptionPaidEvent{
		RefereeID:         5,
		SubscriptionPrice: 100.0,
		OrderID:           "ORD-006",
	})
	assert.NoError(t, err)
	referralRepo.AssertExpectations(t)
}

func TestRebateService_ProcessSubscriptionPaid_UsesDefaultRate(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := &svcconfig.CreditConfig{
		DefaultRebateRate:    0.20,
		ReferralLinkTemplate: "https://app.example.com/referral?code=%s",
	}
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	ref := &model.ReferralRelation{
		ID:                       1,
		ReferrerID:               10,
		RefereeID:                5,
		RefereeSubscriptionCount: 1,
	}
	referralRepo.On("GetByRefereeID", mock.Anything, int64(5)).Return(ref, nil)
	creditRepo.On("GetRebateConfig", mock.Anything, 1).Return(nil, nil)

	creditSvc.On("EarnCredits",
		mock.Anything,
		int64(10),
		20.0,
		"EARN_REFERRAL",
		"rebate:ORD-007:10",
		mock.Anything,
	).Return(nil)
	referralRepo.On("IncrementSubscriptionCount", mock.Anything, int64(5)).Return(nil)

	err := svc.ProcessSubscriptionPaid(context.Background(), &model.ProcessSubscriptionPaidEvent{
		RefereeID:         5,
		SubscriptionPrice: 100.0,
		OrderID:           "ORD-007",
	})
	assert.NoError(t, err)
	creditSvc.AssertExpectations(t)
}

func TestRebateService_GetRebateRate_WithConfig(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	creditRepo.On("GetRebateConfig", mock.Anything, 5).
		Return(&model.RebateConfig{RebatePercentage: 0.12}, nil)

	rate, err := svc.GetRebateRate(context.Background(), 5)
	assert.NoError(t, err)
	assert.Equal(t, 0.12, rate)
	creditRepo.AssertExpectations(t)
}

func TestRebateService_GetRebateRate_WithoutConfig_Default(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	creditRepo.On("GetRebateConfig", mock.Anything, 0).Return(nil, nil)

	rate, err := svc.GetRebateRate(context.Background(), 0)
	assert.NoError(t, err)
	assert.Equal(t, 0.10, rate)
	creditRepo.AssertExpectations(t)
}

func TestRebateService_GetRebateRate_RepoError(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	creditRepo.On("GetRebateConfig", mock.Anything, 5).
		Return(nil, errors.New("db error"))

	rate, err := svc.GetRebateRate(context.Background(), 5)
	assert.Error(t, err)
	assert.Equal(t, 0.0, rate)
	creditRepo.AssertExpectations(t)
}

func TestRebateService_ProcessSubscriptionPaid_RewardAmountCalculation(t *testing.T) {
	creditRepo := new(mockCreditRepo)
	referralRepo := new(mockReferralRepo)
	creditSvc := new(mockCreditService)
	cfg := defaultCreditConfig()
	svc := NewRebateService(creditRepo, referralRepo, creditSvc, cfg)

	ref := &model.ReferralRelation{
		ID:                       1,
		ReferrerID:               20,
		RefereeID:                8,
		RefereeSubscriptionCount: 10,
	}
	referralRepo.On("GetByRefereeID", mock.Anything, int64(8)).Return(ref, nil)
	creditRepo.On("GetRebateConfig", mock.Anything, 10).
		Return(&model.RebateConfig{RebatePercentage: 0.25}, nil)

	expectedReward := 49.98 * 0.25
	creditSvc.On("EarnCredits",
		mock.Anything,
		int64(20),
		expectedReward,
		"EARN_REFERRAL",
		"rebate:ORD-010:20",
		mock.Anything,
	).Return(nil)
	referralRepo.On("IncrementSubscriptionCount", mock.Anything, int64(8)).Return(nil)

	err := svc.ProcessSubscriptionPaid(context.Background(), &model.ProcessSubscriptionPaidEvent{
		RefereeID:         8,
		SubscriptionPrice: 49.98,
		OrderID:           "ORD-010",
	})
	assert.NoError(t, err)
	creditSvc.AssertExpectations(t)
}
