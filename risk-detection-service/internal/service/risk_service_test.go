package service

import (
	"testing"

	"github.com/trigold786/92-Account-Center/risk-detection-service/internal/model"
)

func TestCalculateRiskScore_NoFactors(t *testing.T) {
	svc := &RiskService{}
	score := svc.CalculateRiskScore(RiskFactors{})
	if score != 0 {
		t.Errorf("expected 0 for no factors, got %d", score)
	}
}

func TestCalculateRiskScore_SingleFactor(t *testing.T) {
	svc := &RiskService{}
	factors := RiskFactors{
		{Type: "test", Score: 40, Weight: 2},
	}
	score := svc.CalculateRiskScore(factors)
	if score != 40 {
		t.Errorf("expected 40, got %d", score)
	}
}

func TestCalculateRiskScore_MultipleFactors(t *testing.T) {
	svc := &RiskService{}
	factors := RiskFactors{
		{Type: "geo", Score: 50, Weight: 3},
		{Type: "device", Score: 40, Weight: 2},
	}
	score := svc.CalculateRiskScore(factors)
	if score <= 0 {
		t.Error("score should be positive")
	}
	weightedSum := float64(50*3 + 40*2)
	totalWeight := float64(3 + 2)
	baseScore := int(weightedSum / totalWeight)
	multiplier := 1.0 + float64(len(factors)-1)*0.1
	expectedScore := baseScore * int(multiplier)
	if score != expectedScore {
		t.Errorf("expected %d, got %d", expectedScore, score)
	}
}

func TestCalculateRiskScore_CappedAt100(t *testing.T) {
	svc := &RiskService{}
	factors := RiskFactors{
		{Type: "geo", Score: 90, Weight: 3},
		{Type: "device", Score: 80, Weight: 2},
		{Type: "velocity", Score: 70, Weight: 2},
		{Type: "extra", Score: 60, Weight: 1},
	}
	score := svc.CalculateRiskScore(factors)
	if score > 100 {
		t.Errorf("expected score capped at 100, got %d", score)
	}
}

func TestCalculateRiskLevel_Low(t *testing.T) {
	svc := &RiskService{}
	tests := []int{0, 10, 20, 30}
	for _, score := range tests {
		level := svc.calculateRiskLevel(score)
		if level != model.RiskLevelLow {
			t.Errorf("calculateRiskLevel(%d) = %s, want %s", score, level, model.RiskLevelLow)
		}
	}
}

func TestCalculateRiskLevel_Medium(t *testing.T) {
	svc := &RiskService{}
	tests := []int{31, 40, 50, 60}
	for _, score := range tests {
		level := svc.calculateRiskLevel(score)
		if level != model.RiskLevelMedium {
			t.Errorf("calculateRiskLevel(%d) = %s, want %s", score, level, model.RiskLevelMedium)
		}
	}
}

func TestCalculateRiskLevel_High(t *testing.T) {
	svc := &RiskService{}
	tests := []int{61, 70, 80}
	for _, score := range tests {
		level := svc.calculateRiskLevel(score)
		if level != model.RiskLevelHigh {
			t.Errorf("calculateRiskLevel(%d) = %s, want %s", score, level, model.RiskLevelHigh)
		}
	}
}

func TestCalculateRiskLevel_Critical(t *testing.T) {
	svc := &RiskService{}
	tests := []int{81, 90, 100}
	for _, score := range tests {
		level := svc.calculateRiskLevel(score)
		if level != model.RiskLevelCritical {
			t.Errorf("calculateRiskLevel(%d) = %s, want %s", score, level, model.RiskLevelCritical)
		}
	}
}

func TestDetermineAction_Allow(t *testing.T) {
	svc := &RiskService{}
	tests := []int{0, 15, 30}
	for _, score := range tests {
		action := svc.determineAction(score)
		if action != "allow" {
			t.Errorf("determineAction(%d) = %s, want allow", score, action)
		}
	}
}

func TestDetermineAction_Verify(t *testing.T) {
	svc := &RiskService{}
	tests := []int{31, 50, 80}
	for _, score := range tests {
		action := svc.determineAction(score)
		if action != "verify" {
			t.Errorf("determineAction(%d) = %s, want verify", score, action)
		}
	}
}

func TestDetermineAction_Deny(t *testing.T) {
	svc := &RiskService{}
	tests := []int{81, 95, 100}
	for _, score := range tests {
		action := svc.determineAction(score)
		if action != "deny" {
			t.Errorf("determineAction(%d) = %s, want deny", score, action)
		}
	}
}

func TestCalculateFingerprintSimilarity_SameString(t *testing.T) {
	svc := &RiskService{}
	sim := svc.calculateFingerprintSimilarity("abc123", "abc123")
	if sim != 1.0 {
		t.Errorf("expected 1.0 for same strings, got %f", sim)
	}
}

func TestCalculateFingerprintSimilarity_DifferentStrings(t *testing.T) {
	svc := &RiskService{}
	sim := svc.calculateFingerprintSimilarity("abc123", "xyz789")
	if sim >= 1.0 {
		t.Errorf("expected < 1.0 for different strings, got %f", sim)
	}
	if sim < 0 {
		t.Errorf("similarity should not be negative, got %f", sim)
	}
}

func TestCalculateFingerprintSimilarity_EmptyStrings(t *testing.T) {
	svc := &RiskService{}
	sim := svc.calculateFingerprintSimilarity("", "")
	if sim != 1.0 {
		t.Errorf("expected 1.0 for empty strings, got %f", sim)
	}
}
