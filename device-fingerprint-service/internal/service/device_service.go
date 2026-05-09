package service

import (
	"context"
	"time"

	"device-fingerprint-service/internal/model"
)

// DeviceFingerprintRepository defines the interface for device fingerprint data access
type DeviceFingerprintRepository interface {
	Save(ctx context.Context, fp *model.DeviceFingerprint) error
	GetByFingerprintID(ctx context.Context, userID uint64, fingerprintID string) (*model.DeviceFingerprint, error)
	GetByUserID(ctx context.Context, userID uint64) ([]*model.DeviceFingerprint, error)
	Update(ctx context.Context, fp *model.DeviceFingerprint) error
}

// DeviceFingerprintService handles device fingerprint operations
type DeviceFingerprintService struct {
	repo         DeviceFingerprintRepository
	trustDays    int // Default trust duration in days
	riskThreshold float64 // Risk threshold for feature changes (0.0-1.0)
}

// NewDeviceFingerprintService creates a new device fingerprint service
func NewDeviceFingerprintService(repo DeviceFingerprintRepository, trustDays int, riskThreshold float64) *DeviceFingerprintService {
	if trustDays <= 0 {
		trustDays = 3 // Default to 3 days
	}
	if riskThreshold < 0.0 || riskThreshold > 1.0 {
		riskThreshold = 0.3 // Default to 30% change threshold
	}
	return &DeviceFingerprintService{
		repo:         repo,
		trustDays:    trustDays,
		riskThreshold: riskThreshold,
	}
}

// RegisterDevice registers or updates a device fingerprint for a user
func (s *DeviceFingerprintService) RegisterDevice(ctx context.Context, userID uint64, req *model.DeviceFingerprintRequest) (*model.DeviceFingerprintResponse, error) {
	// Check if device already exists for this user
	existing, err := s.repo.GetByFingerprintID(ctx, userID, req.FingerprintID)
	if err != nil {
		// If error is not "not found", return it
		// We'll assume our repository returns a specific error for not found
		// For now, we'll treat any error as not found and create new
		// In a real implementation, we'd check for specific not found error
		existing = nil
	}

	var device *model.DeviceFingerprint
	if existing != nil {
		device = existing
		// Update existing device
		device.UserAgent = req.UserAgent
		device.IPAddress = req.IPAddress
		device.Country = req.Country
		device.City = req.City
		device.Latitude = req.Latitude
		device.Longitude = req.Longitude
		device.Features = req.Features
		device.UpdatedAt = time.Now().Unix()
	} else {
		// Create new device
		device = &model.DeviceFingerprint{
			UserID:       userID,
			FingerprintID: req.FingerprintID,
			UserAgent:    req.UserAgent,
			IPAddress:    req.IPAddress,
			Country:      req.Country,
			City:         req.City,
			Latitude:     req.Latitude,
			Longitude:    req.Longitude,
			Features:     req.Features,
			IsTrusted:    false, // New devices are not trusted by default
			LastUsedAt:   time.Now().Unix(),
			CreatedAt:    time.Now().Unix(),
			UpdatedAt:    time.Now().Unix(),
		}
	}

	// Save device
	if err := s.repo.Save(ctx, device); err != nil {
		return nil, err
	}

	// Check if device should be trusted based on history
	// In a real implementation, we might check if user has recently passed strong authentication
	// For now, we'll leave IsTrusted as false and let the auth service set it after successful verification
	return &model.DeviceFingerprintResponse{
		ID:          device.ID,
		FingerprintID: device.FingerprintID,
		IsTrusted:   device.IsTrusted,
		LastUsedAt:  device.LastUsedAt,
	}, nil
}

// GetDevice gets a device fingerprint by fingerprint ID for a user
func (s *DeviceFingerprintService) GetDevice(ctx context.Context, userID uint64, fingerprintID string) (*model.DeviceFingerprint, error) {
	return s.repo.GetByFingerprintID(ctx, userID, fingerprintID)
}

// ListDevices lists all devices for a user
func (s *DeviceFingerprintService) ListDevices(ctx context.Context, userID uint64) ([]*model.DeviceFingerprint, error) {
	return s.repo.GetByUserID(ctx, userID)
}

// IsTrusted checks if a device is trusted for a user
// A device is trusted if it was used within the last N days (where N is trustDays)
// and the user has passed strong verification recently (handled by auth service)
func (s *DeviceFingerprintService) IsTrusted(ctx context.Context, userID uint64, fingerprintID string) (bool, error) {
	device, err := s.repo.GetByFingerprintID(ctx, userID, fingerprintID)
	if err != nil {
		return false, err
	}

	if device == nil {
		return false, nil
	}

	// Check if device is within trust period
	now := time.Now().Unix()
	lastUsed := time.Unix(device.LastUsedAt, 0)
	if now.Sub(lastUsed) > time.Duration(s.trustDays)*24*time.Hour {
		return false, nil
	}

	return device.IsTrusted, nil
}

// AssessRisk evaluates the risk level of a device fingerprint
// Returns true if risk is high (should trigger re-verification)
func (s *DeviceFingerprintService) AssessRisk(ctx context.Context, userID uint64, req *model.DeviceFingerprintRequest) (bool, error) {
	// Get existing device for this user and fingerprint
	existing, err := s.repo.GetByFingerprintID(ctx, userID, req.FingerprintID)
	if err != nil {
		// If device doesn't exist, treat as high risk (new device)
		if err == model.ErrDeviceNotFound { // We'll define this error in model package
			return true, nil
		}
		return false, err
	}

	if existing == nil {
		return true, nil // New device is high risk
	}

	// Check geolocation risk
	geoRisk := s.isGeoLocationRisky(existing, req)
	if geoRisk {
		return true, nil
	}

	// Check device feature risk
	featureRisk := s.isFeatureRisky(existing.Features, req.Features)
	if featureRisk {
		return true, nil
	}

	return false, nil
}

// isGeoLocationRisky checks if geolocation change is drastic
func (s *DeviceFingerprintService) isGeoLocationRisky(existing *model.DeviceFingerprint, req *model.DeviceFingerprintRequest) bool {
	// Simple distance calculation using Haversine formula would be ideal
	// For simplicity, we'll check if country or city changed significantly
	// In a real implementation, we'd use a proper geolocation library
	if existing.Country != req.Country {
		return true // Country change is considered risky
	}
	// For city, we could check distance, but for now just check if different
	if existing.City != req.City && req.City != "" {
		return true
	}
	// TODO: Implement proper distance check using latitude/longitude
	return false
}

// isFeatureRisky checks if device feature changes exceed threshold
func (s *DeviceFingerprintService) isFeatureRisky(existingFeatures, newFeatures []byte) bool {
	// In a real implementation, we'd compare the actual device features
	// For now, we'll do a simple byte comparison and calculate change ratio
	if len(existingFeatures) == 0 || len(newFeatures) == 0 {
		return true // If either is empty, treat as risky
	}

	// Simple comparison: count different bytes
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
	// Account for length difference
	diffCount += abs(len(existingFeatures) - len(newFeatures))

	changeRatio := float64(diffCount) / float64(max(len(existingFeatures), len(newFeatures)))
	return changeRatio > s.riskThreshold
}

// Helper functions
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}