package model

import "time"

type AssessRiskRequest struct {
	UserID            string    `json:"user_id" binding:"required"`
	IPAddress         string    `json:"ip_address" binding:"required"`
	DeviceFingerprint string    `json:"device_fingerprint" binding:"required"`
	Timestamp         time.Time `json:"timestamp" binding:"required"`
	Location          *Location `json:"location,omitempty"`
}

type RiskFactor struct {
	Type     string  `json:"type"`
	Score    int     `json:"score"`
	Weight   int     `json:"weight"`
	Detail   string  `json:"detail"`
	Location *Location `json:"location,omitempty"`
}

type RiskAssessmentResponse struct {
	RiskScore   int          `json:"risk_score"`
	RiskLevel   RiskLevel    `json:"risk_level"`
	RiskFactors []RiskFactor `json:"risk_factors"`
	Action      string       `json:"action"`
}

type GetRiskHistoryRequest struct {
	UserID    string    `json:"user_id" binding:"required"`
	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	Limit     int       `json:"limit"`
}