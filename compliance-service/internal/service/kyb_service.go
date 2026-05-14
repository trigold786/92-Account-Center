package service

import (
	"context"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"math/big"
	"time"

	"github.com/google/uuid"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/repository"
	"github.com/trigold786/92-Account-Center/compliance-service/pkg/crypto"
)

var (
	ErrEnterpriseNotFound      = errors.New("enterprise not found")
	ErrInvalidAmount           = errors.New("invalid micro payment amount")
	ErrMicroPaymentNotPending  = errors.New("micro payment is not in pending status")
	ErrFaceVerificationFailed  = errors.New("face verification failed")
	ErrEncryptionFailed        = errors.New("encryption failed")
)

type KYBService interface {
	SubmitEnterpriseInfo(ctx context.Context, userID string, req *model.EnterpriseInfoRequest) (*model.EnterpriseSubmitResponse, error)
	InitiateMicroPayment(ctx context.Context, enterpriseID string) (*model.MicroPaymentInitResponse, error)
	VerifyMicroPayment(ctx context.Context, enterpriseID string, amount float64) error
	SubmitFaceVerification(ctx context.Context, enterpriseID string, token string) error
	GetEnterpriseStatus(ctx context.Context, enterpriseID string) (*model.EnterpriseStatusResponse, error)
}

type kybService struct {
	repo       repository.EnterpriseRepository
	encryptKey []byte
}

func NewKYBService(repo repository.EnterpriseRepository, encryptKey []byte) KYBService {
	return &kybService{
		repo:       repo,
		encryptKey: encryptKey,
	}
}

func (s *kybService) SubmitEnterpriseInfo(ctx context.Context, userID string, req *model.EnterpriseInfoRequest) (*model.EnterpriseSubmitResponse, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	encryptedIDNumber, err := s.encrypt(req.LegalPersonIDNumber)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	encryptedBankAccount, err := s.encrypt(req.BankAccountNumber)
	if err != nil {
		return nil, ErrEncryptionFailed
	}

	enterpriseID := uuid.New()
	now := time.Now()

	enterprise := &model.Enterprise{
		EnterpriseID:            enterpriseID,
		UserID:                  uid,
		CompanyName:             req.CompanyName,
		UnifiedSocialCreditCode: req.UnifiedSocialCreditCode,
		LegalPersonName:         req.LegalPersonName,
		LegalPersonIDNumber:     encryptedIDNumber,
		BankName:                req.BankName,
		BankAccountNumber:       encryptedBankAccount,
		VerificationStatus:      model.VerificationStatusPending,
		MicroPaymentStatus:      model.MicroPaymentStatusPending,
		FaceVerificationStatus:  model.FaceVerificationStatusPending,
		CreatedAt:               now,
		UpdatedAt:               now,
	}

	if err := s.repo.Create(ctx, enterprise); err != nil {
		return nil, err
	}

	return &model.EnterpriseSubmitResponse{
		EnterpriseID:       enterpriseID.String(),
		VerificationStatus: string(model.VerificationStatusPending),
		Message:            "Enterprise info submitted successfully, pending micro payment verification",
	}, nil
}

func (s *kybService) InitiateMicroPayment(ctx context.Context, enterpriseID string) (*model.MicroPaymentInitResponse, error) {
	eid, err := uuid.Parse(enterpriseID)
	if err != nil {
		return nil, errors.New("invalid enterprise ID")
	}

	enterprise, err := s.repo.GetByID(ctx, eid)
	if err != nil {
		return nil, err
	}
	if enterprise == nil {
		return nil, ErrEnterpriseNotFound
	}

	if enterprise.MicroPaymentStatus != model.MicroPaymentStatusPending {
		return nil, ErrMicroPaymentNotPending
	}

	amount, err := generateMicroPaymentAmount()
	if err != nil {
		return nil, err
	}

	enterprise.MicroPaymentAmount = amount
	enterprise.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, enterprise); err != nil {
		return nil, err
	}

	maskedAccount := maskBankAccount(enterprise.BankAccountNumber)

	return &model.MicroPaymentInitResponse{
		EnterpriseID:    enterpriseID,
		Amount:          amount,
		BankName:       enterprise.BankName,
		BankAccountMask: maskedAccount,
	}, nil
}

func (s *kybService) VerifyMicroPayment(ctx context.Context, enterpriseID string, amount float64) error {
	eid, err := uuid.Parse(enterpriseID)
	if err != nil {
		return errors.New("invalid enterprise ID")
	}

	enterprise, err := s.repo.GetByID(ctx, eid)
	if err != nil {
		return err
	}
	if enterprise == nil {
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
		return s.repo.Update(ctx, enterprise)
	}

	enterprise.MicroPaymentStatus = model.MicroPaymentStatusVerified
	enterprise.UpdatedAt = time.Now()

	s.updateOverallVerificationStatus(enterprise)

	return s.repo.Update(ctx, enterprise)
}

func (s *kybService) SubmitFaceVerification(ctx context.Context, enterpriseID string, token string) error {
	eid, err := uuid.Parse(enterpriseID)
	if err != nil {
		return errors.New("invalid enterprise ID")
	}

	enterprise, err := s.repo.GetByID(ctx, eid)
	if err != nil {
		return err
	}
	if enterprise == nil {
		return ErrEnterpriseNotFound
	}

	if token == "" {
		enterprise.FaceVerificationStatus = model.FaceVerificationStatusFailed
		enterprise.UpdatedAt = time.Now()
		return s.repo.Update(ctx, enterprise)
	}

	score, err := generateFaceScore()
	if err != nil {
		return err
	}

	if score < 0.8 {
		enterprise.FaceVerificationStatus = model.FaceVerificationStatusFailed
		enterprise.UpdatedAt = time.Now()
		return s.repo.Update(ctx, enterprise)
	}

	enterprise.FaceVerificationScore = score
	enterprise.FaceVerificationStatus = model.FaceVerificationStatusVerified
	enterprise.UpdatedAt = time.Now()

	s.updateOverallVerificationStatus(enterprise)

	return s.repo.Update(ctx, enterprise)
}

func (s *kybService) GetEnterpriseStatus(ctx context.Context, enterpriseID string) (*model.EnterpriseStatusResponse, error) {
	eid, err := uuid.Parse(enterpriseID)
	if err != nil {
		return nil, errors.New("invalid enterprise ID")
	}

	enterprise, err := s.repo.GetByID(ctx, eid)
	if err != nil {
		return nil, err
	}
	if enterprise == nil {
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

func (s *kybService) encrypt(plaintext string) (string, error) {
	if s.encryptKey == nil || len(s.encryptKey) != 16 {
		key := make([]byte, 16)
		if _, err := rand.Read(key); err != nil {
			return "", err
		}
		s.encryptKey = key
	}

	block, err := crypto.NewCipher(s.encryptKey)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func maskBankAccount(account string) string {
	if len(account) <= 4 {
		return account
	}
	return "****" + account[len(account)-4:]
}

func generateMicroPaymentAmount() (float64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(90))
	if err != nil {
		return 0, err
	}
	amount := 0.01 + float64(n.Int64())/1000.0
	return float64(int(amount*100)) / 100.0, nil
}

func generateFaceScore() (float64, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(30))
	if err != nil {
		return 0, err
	}
	score := 0.70 + float64(n.Int64())/100.0
	return float64(int(score*100)) / 100.0, nil
}
