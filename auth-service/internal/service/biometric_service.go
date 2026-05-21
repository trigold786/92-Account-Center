package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

var ErrBiometricTokenExpired = errors.New("biometric token expired")

type BiometricRepository interface {
	GetDeviceToken(ctx context.Context, userID int64, deviceID string) (string, error)
	SaveDeviceToken(ctx context.Context, userID int64, deviceID, token string, expiry time.Duration) error
	DeleteDeviceToken(ctx context.Context, userID int64, deviceID string) error
}

type BiometricService struct {
	repo BiometricRepository
}

func NewBiometricService(repo BiometricRepository) *BiometricService {
	return &BiometricService{repo: repo}
}

func (s *BiometricService) BindDeviceToken(ctx context.Context, userID int64, deviceID, token string) error {
	return s.repo.SaveDeviceToken(ctx, userID, deviceID, token, 90*24*time.Hour)
}

func (s *BiometricService) ValidateToken(ctx context.Context, userID int64, deviceID, token string) (bool, error) {
	stored, err := s.repo.GetDeviceToken(ctx, userID, deviceID)
	if err != nil {
		return false, err
	}
	if stored == "" || stored != token {
		return false, nil
	}
	return true, nil
}

func (s *BiometricService) RotateToken(ctx context.Context, userID int64, oldToken string) (string, error) {
	b := make([]byte, 32)
	rand.Read(b)
	newToken := hex.EncodeToString(b)
	return newToken, nil
}

func (s *BiometricService) RevokeToken(ctx context.Context, userID int64, deviceID string) error {
	return s.repo.DeleteDeviceToken(ctx, userID, deviceID)
}
