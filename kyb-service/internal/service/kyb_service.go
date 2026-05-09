package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/google/uuid"

	"account-center/kyb-service/internal/model"
)

var (
	ErrEnterpriseNotFound    = errors.New("enterprise not found")
	ErrInvalidAmount         = errors.New("invalid micro payment amount")
	ErrMicroPaymentNotPending = errors.New("micro payment is not in pending status")
	ErrFaceVerificationFailed = errors.New("face verification failed")
)

type KYBService interface {
	SubmitEnterpriseInfo(ctx context.Context, userID string, req *model.EnterpriseInfoRequest) (*model.EnterpriseSubmitResponse, error)
	InitiateMicroPayment(ctx context.Context, enterpriseID string) (*model.MicroPaymentInitResponse, error)
	VerifyMicroPayment(ctx context.Context, enterpriseID string, amount float64) error
	SubmitFaceVerification(ctx context.Context, enterpriseID string, token string) error
	GetEnterpriseStatus(ctx context.Context, enterpriseID string) (*model.EnterpriseStatusResponse, error)
}

type kybService struct {
	enterprises map[string]*model.Enterprise
}

func NewKYBService() KYBService {
	return &kybService{
		enterprises: make(map[string]*model.Enterprise),
	}
}

func (s *kybService) SubmitEnterpriseInfo(ctx context.Context, userID string, req *model.EnterpriseInfoRequest) (*model.EnterpriseSubmitResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	enterpriseID := uuid.New()
	now := time.Now()

	enterprise := &model.Enterprise{
		EnterpriseID:           enterpriseID,
		UserID:                 uid,
		CompanyName:            req.CompanyName,
		UnifiedSocialCreditCode: req.UnifiedSocialCreditCode,
		LegalPersonName:        req.LegalPersonName,
		LegalPersonIDNumber:    req.LegalPersonIDNumber,
		BankName:               req.BankName,
		BankAccountNumber:      req.BankAccountNumber,
		VerificationStatus:     model.VerificationStatusPending,
		MicroPaymentStatus:     model.MicroPaymentStatusPending,
		FaceVerificationStatus: model.FaceVerificationStatusPending,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	s.enterprises[enterpriseID.String()] = enterprise

	return &model.EnterpriseSubmitResponse{
		EnterpriseID:      enterpriseID.String(),
		VerificationStatus: string(model.VerificationStatusPending),
		Message:            "Enterprise info submitted successfully, pending micro payment verification",
	}, nil
}

func (s *kybService) InitiateMicroPayment(ctx context.Context, enterpriseID string) (*model.MicroPaymentInitResponse, error) {
	enterprise, ok := s.enterprises[enterpriseID]
	if !ok {
		return nil, ErrEnterpriseNotFound
	}

	if enterprise.MicroPaymentStatus != model.MicroPaymentStatusPending {
		return nil, ErrMicroPaymentNotPending
	}

	amount := 0.01 + rand.Float64()*0.09
	amount = float64(int(amount*100)) / 100.0

	enterprise.MicroPaymentAmount = amount
	enterprise.UpdatedAt = time.Now()

	maskedAccount := maskBankAccount(enterprise.BankAccountNumber)

	return &model.MicroPaymentInitResponse{
		EnterpriseID:    enterpriseID,
		Amount:         amount,
		BankName:       enterprise.BankName,
		BankAccountMask: maskedAccount,
	}, nil
}

func (s *kybService) VerifyMicroPayment(ctx context.Context, enterpriseID string, amount float64) error {
	enterprise, ok := s.enterprises[enterpriseID]
	if !ok {
		return ErrEnterpriseNotFound
	}

	if enterprise.MicroPaymentStatus != model.MicroPaymentStatusPending {
		return ErrMicroPaymentNotPending
	}

	if amount < 0.01 || amount > 0.1 {
		return ErrInvalidAmount
	}

	allowedRange := enterprise.MicroPaymentAmount * 1.1
	lowerRange := enterprise.MicroPaymentAmount * 0.9

	if amount < lowerRange || amount > allowedRange {
		enterprise.MicroPaymentStatus = model.MicroPaymentStatusFailed
		enterprise.UpdatedAt = time.Now()
		return errors.New("micro payment amount verification failed")
	}

	enterprise.MicroPaymentStatus = model.MicroPaymentStatusVerified
	enterprise.UpdatedAt = time.Now()

	s.updateOverallVerificationStatus(enterprise)

	return nil
}

func (s *kybService) SubmitFaceVerification(ctx context.Context, enterpriseID string, token string) error {
	enterprise, ok := s.enterprises[enterpriseID]
	if !ok {
		return ErrEnterpriseNotFound
	}

	if token == "" {
		enterprise.FaceVerificationStatus = model.FaceVerificationStatusFailed
		enterprise.UpdatedAt = time.Now()
		return ErrFaceVerificationFailed
	}

	score := 0.7 + rand.Float64()*0.3
	score = float64(int(score*100)) / 100.0

	if score < 0.8 {
		enterprise.FaceVerificationStatus = model.FaceVerificationStatusFailed
		enterprise.UpdatedAt = time.Now()
		return errors.New("face verification score too low")
	}

	enterprise.FaceVerificationScore = score
	enterprise.FaceVerificationStatus = model.FaceVerificationStatusVerified
	enterprise.UpdatedAt = time.Now()

	s.updateOverallVerificationStatus(enterprise)

	return nil
}

func (s *kybService) GetEnterpriseStatus(ctx context.Context, enterpriseID string) (*model.EnterpriseStatusResponse, error) {
	enterprise, ok := s.enterprises[enterpriseID]
	if !ok {
		return nil, ErrEnterpriseNotFound
	}

	return &model.EnterpriseStatusResponse{
		EnterpriseID:           enterprise.EnterpriseID.String(),
		CompanyName:            enterprise.CompanyName,
		VerificationStatus:     string(enterprise.VerificationStatus),
		MicroPaymentStatus:     string(enterprise.MicroPaymentStatus),
		MicroPaymentAmount:     enterprise.MicroPaymentAmount,
		FaceVerificationStatus: string(enterprise.FaceVerificationStatus),
		FaceVerificationScore:  enterprise.FaceVerificationScore,
		CreatedAt:              enterprise.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              enterprise.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *kybService) updateOverallVerificationStatus(enterprise *model.Enterprise) {
	if enterprise.MicroPaymentStatus == model.MicroPaymentStatusVerified &&
		enterprise.FaceVerificationStatus == model.FaceVerificationStatusVerified {
		enterprise.VerificationStatus = model.VerificationStatusVerified
	} else if enterprise.MicroPaymentStatus == model.MicroPaymentStatusFailed ||
		enterprise.FaceVerificationStatus == model.FaceVerificationStatusFailed {
		enterprise.VerificationStatus = model.VerificationStatusFailed
	} else {
		enterprise.VerificationStatus = model.VerificationStatusPending
	}
}

func maskBankAccount(account string) string {
	if len(account) <= 4 {
		return account
	}
	return "****" + account[len(account)-4:]
}