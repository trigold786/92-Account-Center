package service

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/repository"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/svcconfig"
)

const (
	impossibleTravelDistance = 1000.0
	impossibleTravelHours   = 1.0
	deviceChangeThreshold   = 0.5
	velocityWindowHours     = 1.0
)

type RiskService struct {
	repo       *repository.RiskRepository
	geoService *GeoService
	cfg        *svcconfig.ComplianceConfig
}

func NewRiskService(repo *repository.RiskRepository, geoService *GeoService, cfg *svcconfig.ComplianceConfig) *RiskService {
	return &RiskService{
		repo:       repo,
		geoService: geoService,
		cfg:        cfg,
	}
}

type RiskFactors []model.RiskFactor

func (s *RiskService) AssessRisk(ctx context.Context, req model.AssessRiskRequest) (*model.RiskAssessmentResponse, error) {
	var factors RiskFactors

	location := req.Location
	if location == nil {
		location = &model.Location{}
	}

	geoAnomaly, distance, err := s.DetectGeoAnomaly(ctx, req.UserID, location)
	if err != nil {
		return nil, err
	}
	if geoAnomaly {
		factors = append(factors, model.RiskFactor{
			Type:   "impossible_travel",
			Score:  50,
			Weight: 3,
			Detail: "Impossible travel detected",
		})
	}

	deviceAnomaly, _, err := s.DetectDeviceAnomaly(ctx, req.UserID, req.DeviceFingerprint)
	if err != nil {
		return nil, err
	}
	if deviceAnomaly {
		factors = append(factors, model.RiskFactor{
			Type:   "device_change",
			Score:  40,
			Weight: 2,
			Detail: "Significant device fingerprint change",
		})
	}

	velocityAnomaly, _, err := s.DetectVelocityAnomaly(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if velocityAnomaly {
		factors = append(factors, model.RiskFactor{
			Type:   "velocity_exceeded",
			Score:  30,
			Weight: 2,
			Detail: "Abnormal login frequency detected",
		})
	}

	riskScore := s.CalculateRiskScore(factors)

	riskLevel := s.calculateRiskLevel(riskScore)
	action := s.determineAction(riskScore)

	event := model.RiskEvent{
		RiskEventID: uuid.New().String(),
		UserID:      req.UserID,
		EventType:   model.EventTypeLogin,
		RiskScore:   riskScore,
		RiskLevel:   riskLevel,
		IPAddress:   req.IPAddress,
		CreatedAt:   time.Now(),
	}

	detailsJSON, _ := json.Marshal(map[string]interface{}{
		"factors":     factors,
		"distance_km": distance,
	})
	event.Details = detailsJSON

	s.RecordRiskEvent(ctx, &event)

	return &model.RiskAssessmentResponse{
		RiskScore:   riskScore,
		RiskLevel:   riskLevel,
		RiskFactors: factors,
		Action:      action,
	}, nil
}

func (s *RiskService) DetectGeoAnomaly(ctx context.Context, userID string, newLocation *model.Location) (bool, float64, error) {
	if userID == "" || newLocation == nil {
		return false, 0, nil
	}

	lastEvent, err := s.repo.GetLastEventByUserID(ctx, userID)
	if err != nil || lastEvent == nil || lastEvent.Location == nil {
		return false, 0, err
	}

	distance := s.geoService.CalculateDistance(
		lastEvent.Location.Latitude, lastEvent.Location.Longitude,
		newLocation.Latitude, newLocation.Longitude,
	)

	if distance > impossibleTravelDistance {
		timeDelta := time.Since(lastEvent.CreatedAt).Hours()
		if timeDelta < impossibleTravelHours {
			return true, distance, nil
		}
	}

	return false, distance, nil
}

func (s *RiskService) DetectDeviceAnomaly(ctx context.Context, userID, newFingerprint string) (bool, float64, error) {
	if userID == "" || newFingerprint == "" {
		return false, 0, nil
	}

	lastEvent, err := s.repo.GetLastEventByUserID(ctx, userID)
	if err != nil || lastEvent == nil {
		return false, 0, err
	}

	details := make(map[string]interface{})
	if err := json.Unmarshal(lastEvent.Details, &details); err != nil {
		return false, 0, err
	}

	lastFingerprint, ok := details["device_fingerprint"].(string)
	if !ok {
		return false, 0, nil
	}

	similarity := s.calculateFingerprintSimilarity(lastFingerprint, newFingerprint)
	changeRate := 1.0 - similarity

	return changeRate > deviceChangeThreshold, similarity, nil
}

func (s *RiskService) DetectVelocityAnomaly(ctx context.Context, userID string) (bool, int, error) {
	if userID == "" {
		return false, 0, nil
	}

	windowStart := time.Now().Add(-time.Duration(velocityWindowHours) * time.Hour)
	count, err := s.repo.CountEventsByUserIDSince(ctx, userID, windowStart)
	if err != nil {
		return false, 0, err
	}

	return count >= s.cfg.RiskRegistrationRateLimit, int(count), nil
}

func (s *RiskService) CalculateRiskScore(factors RiskFactors) int {
	if len(factors) == 0 {
		return 0
	}

	var totalWeight float64
	var weightedSum float64

	for _, f := range factors {
		totalWeight += float64(f.Weight)
		weightedSum += float64(f.Score) * float64(f.Weight)
	}

	rawScore := weightedSum / totalWeight
	baseScore := int(rawScore)

	multiplier := 1.0 + float64(len(factors)-1)*0.1
	finalScore := baseScore * int(multiplier)

	if finalScore > s.cfg.RiskMaxScore {
		finalScore = s.cfg.RiskMaxScore
	}

	return finalScore
}

func (s *RiskService) calculateRiskLevel(score int) model.RiskLevel {
	switch {
	case score <= 30:
		return model.RiskLevelLow
	case score <= 60:
		return model.RiskLevelMedium
	case score <= 80:
		return model.RiskLevelHigh
	default:
		return model.RiskLevelCritical
	}
}

func (s *RiskService) determineAction(score int) string {
	switch {
	case score <= 30:
		return "allow"
	case score <= 80:
		return "verify"
	default:
		return "deny"
	}
}

func (s *RiskService) calculateFingerprintSimilarity(fp1, fp2 string) float64 {
	if fp1 == fp2 {
		return 1.0
	}

	hash1 := sha256.Sum256([]byte(fp1))
	hash2 := sha256.Sum256([]byte(fp2))

	var matchingBits int
	for i := 0; i < len(hash1); i++ {
		xor := hash1[i] ^ hash2[i]
		for xor > 0 {
			matchingBits += int(xor & 1)
			xor >>= 1
		}
	}

	totalBits := len(hash1) * 8
	return float64(matchingBits) / float64(totalBits)
}

func (s *RiskService) RecordRiskEvent(ctx context.Context, event *model.RiskEvent) error {
	return s.repo.Create(ctx, event)
}

func (s *RiskService) GetRiskHistory(ctx context.Context, userID string, start, end time.Time) ([]*model.RiskEvent, error) {
	return s.repo.GetByUserID(ctx, userID, start, end, s.cfg.RiskHistoryDefaultLimit)
}
