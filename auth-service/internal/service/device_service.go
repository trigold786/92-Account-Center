package service

import (
	"context"
	"errors"
	"time"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
	"github.com/trigold786/92-Account-Center/auth-service/internal/repository"
)

var ErrDeviceNotFound = errors.New("device not found")

type DeviceFingerprintService struct {
	repo          repository.DeviceRepository
	trustDays     int
	riskThreshold float64
}

func NewDeviceFingerprintService(repo repository.DeviceRepository, trustDays int, riskThreshold float64) *DeviceFingerprintService {
	if trustDays <= 0 {
		trustDays = 30
	}
	if riskThreshold < 0.0 || riskThreshold > 1.0 {
		riskThreshold = 0.3
	}
	return &DeviceFingerprintService{
		repo:          repo,
		trustDays:     trustDays,
		riskThreshold: riskThreshold,
	}
}

func (s *DeviceFingerprintService) RegisterDevice(ctx context.Context, userID uint64, req *model.DeviceFingerprintRequest) (*model.DeviceFingerprintResponse, error) {
	existing, _ := s.repo.GetByFingerprintID(ctx, userID, req.FingerprintID)

	var device *model.DeviceFingerprint
	if existing != nil {
		device = existing
		device.UserAgent = req.UserAgent
		device.IPAddress = req.IPAddress
		device.Country = req.Country
		device.City = req.City
		device.Latitude = req.Latitude
		device.Longitude = req.Longitude
		device.Features = req.Features
		device.UpdatedAt = time.Now().Unix()
	} else {
		device = &model.DeviceFingerprint{
			UserID:        userID,
			FingerprintID: req.FingerprintID,
			UserAgent:     req.UserAgent,
			IPAddress:     req.IPAddress,
			Country:       req.Country,
			City:          req.City,
			Latitude:      req.Latitude,
			Longitude:     req.Longitude,
			Features:      req.Features,
			IsTrusted:     false,
			LastUsedAt:    time.Now().Unix(),
			CreatedAt:     time.Now().Unix(),
			UpdatedAt:     time.Now().Unix(),
		}
	}

	if err := s.repo.Save(ctx, device); err != nil {
		return nil, err
	}

	return &model.DeviceFingerprintResponse{
		ID:            device.ID,
		FingerprintID: device.FingerprintID,
		IsTrusted:     device.IsTrusted,
		LastUsedAt:    device.LastUsedAt,
	}, nil
}

func (s *DeviceFingerprintService) VerifyDevice(ctx context.Context, userID uint64, req *model.DeviceFingerprintRequest) (*model.DeviceFingerprintResponse, error) {
	return s.RegisterDevice(ctx, userID, req)
}

func (s *DeviceFingerprintService) TrustDevice(ctx context.Context, userID uint64, fingerprintID string) error {
	device, err := s.repo.GetByFingerprintID(ctx, userID, fingerprintID)
	if err != nil {
		return err
	}
	if device == nil {
		return ErrDeviceNotFound
	}
	device.IsTrusted = true
	device.UpdatedAt = time.Now().Unix()
	return s.repo.Update(ctx, device)
}

func (s *DeviceFingerprintService) GetUserDevices(ctx context.Context, userID uint64) ([]*model.DeviceFingerprint, error) {
	return s.repo.GetByUserID(ctx, userID)
}

func (s *DeviceFingerprintService) RemoveDevice(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}

func (s *DeviceFingerprintService) IsTrusted(ctx context.Context, userID uint64, fingerprintID string) (bool, error) {
	device, err := s.repo.GetByFingerprintID(ctx, userID, fingerprintID)
	if err != nil {
		return false, err
	}
	if device == nil {
		return false, nil
	}

	if !device.IsTrusted {
		return false, nil
	}

	trustExpiry := time.Unix(device.LastUsedAt, 0).Add(time.Duration(s.trustDays) * 24 * time.Hour)
	if time.Now().After(trustExpiry) {
		return false, nil
	}

	return true, nil
}

func (s *DeviceFingerprintService) AssessRisk(ctx context.Context, userID uint64, req *model.DeviceFingerprintRequest) (bool, error) {
	existing, err := s.repo.GetByFingerprintID(ctx, userID, req.FingerprintID)
	if err != nil && err != ErrDeviceNotFound {
		return false, err
	}
	if existing == nil {
		return true, nil
	}

	if s.isGeoLocationRisky(existing, req) {
		return true, nil
	}
	if s.isFeatureRisky(existing.Features, req.Features) {
		return true, nil
	}

	return false, nil
}

func (s *DeviceFingerprintService) isGeoLocationRisky(existing *model.DeviceFingerprint, req *model.DeviceFingerprintRequest) bool {
	if existing.Country != req.Country {
		return true
	}
	if existing.City != req.City && req.City != "" {
		return true
	}
	return false
}

func (s *DeviceFingerprintService) isFeatureRisky(existingFeatures, newFeatures []byte) bool {
	if len(existingFeatures) == 0 || len(newFeatures) == 0 {
		return true
	}

	var diffCount int
	minLen := len(existingFeatures)
	if len(newFeatures) < minLen {
		minLen = len(newFeatures)
	}
	for i := 0; i < minLen; i++ {
		if existingFeatures[i] != newFeatures[i] {
			diffCount++
		}
	}
	diffCount += abs(len(existingFeatures) - len(newFeatures))

	maxLen := len(existingFeatures)
	if len(newFeatures) > maxLen {
		maxLen = len(newFeatures)
	}
	changeRatio := float64(diffCount) / float64(maxLen)
	return changeRatio > s.riskThreshold
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
