package service

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"

	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
	"github.com/trigold786/92-Account-Center/credit-service/internal/repository"
	"github.com/trigold786/92-Account-Center/credit-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/credit-service/pkg/crypto"
)

var (
	ErrInvalidReferralCode = errors.New("invalid referral code")
	ErrAlreadyReferred     = errors.New("user already has a referrer")
	ErrSelfReferral        = errors.New("cannot refer yourself")
)

type ReferralService interface {
	BindReferral(ctx context.Context, referrerCode, refereeID string) error
	GenerateLink(ctx context.Context, userID int64) (*model.GenerateLinkResponse, error)
	GetSummary(ctx context.Context, referrerID int64) (*model.ReferralSummary, error)
}

type referralService struct {
	referralRepo repository.ReferralRepository
	cfg          *svcconfig.CreditConfig
}

func NewReferralService(referralRepo repository.ReferralRepository, cfg *svcconfig.CreditConfig) ReferralService {
	return &referralService{referralRepo: referralRepo, cfg: cfg}
}

func (s *referralService) BindReferral(ctx context.Context, referrerCode, refereeID string) error {
	referrerID, err := decodeReferralCode(referrerCode)
	if err != nil {
		return ErrInvalidReferralCode
	}

	refereeIDInt, err := strconv.ParseInt(refereeID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid referee_id: %w", err)
	}

	if referrerID == refereeIDInt {
		return ErrSelfReferral
	}

	existing, err := s.referralRepo.GetByRefereeID(ctx, refereeIDInt)
	if err != nil {
		return err
	}
	if existing != nil {
		return ErrAlreadyReferred
	}

	_, err = s.referralRepo.Create(ctx, referrerID, refereeIDInt)
	return err
}

func (s *referralService) GenerateLink(ctx context.Context, userID int64) (*model.GenerateLinkResponse, error) {
	code := encodeReferralCode(userID)
	link := fmt.Sprintf(s.cfg.ReferralLinkTemplate, code)
	return &model.GenerateLinkResponse{
		ReferralCode: code,
		ReferralLink: link,
	}, nil
}

func (s *referralService) GetSummary(ctx context.Context, referrerID int64) (*model.ReferralSummary, error) {
	return s.referralRepo.GetReferralSummary(ctx, referrerID)
}

func encodeReferralCode(userID int64) string {
	raw := fmt.Sprintf("%d:%s", userID, crypto.SM3Hash(fmt.Sprintf("referral_%d", userID))[:4])
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func decodeReferralCode(code string) (int64, error) {
	data, err := base64.RawURLEncoding.DecodeString(code)
	if err != nil {
		return 0, ErrInvalidReferralCode
	}
	str := string(data)

	colonIdx := -1
	for i, c := range str {
		if c == ':' {
			colonIdx = i
			break
		}
	}
	if colonIdx < 0 {
		return 0, ErrInvalidReferralCode
	}

	userIDStr := str[:colonIdx]
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		return 0, ErrInvalidReferralCode
	}

	expected := fmt.Sprintf("%d:%s", userID, crypto.SM3Hash(fmt.Sprintf("referral_%d", userID))[:4])
	if str != expected {
		return 0, ErrInvalidReferralCode
	}

	return userID, nil
}
