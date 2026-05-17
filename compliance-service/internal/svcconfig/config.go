package svcconfig

import (
	"fmt"
	"strings"
	"time"

	"github.com/trigold786/92-Account-Center/pkg/config"
)

type ComplianceConfig struct {
	RiskRegistrationRateLimit   int
	RiskMaxScore                int
	KYBFaceScoreThreshold       float64
	BlacklistCacheTTL           time.Duration
	SlidingWindowRegLimit       int
	SlidingWindowRegWindow      time.Duration
	SlidingWindowRefAbuseLimit  int
	SlidingWindowRefAbuseWindow time.Duration
	AuditLogDefaultLimit        int
	RiskHistoryDefaultLimit     int
	EncryptionKey               string
}

func Load(c *config.Client) (*ComplianceConfig, error) {
	cfg := &ComplianceConfig{}
	var errs []string

	if v, err := c.GetConfigInt("RISK_REGISTRATION_RATE_LIMIT"); err != nil {
		errs = append(errs, fmt.Sprintf("RISK_REGISTRATION_RATE_LIMIT: %v", err))
	} else {
		cfg.RiskRegistrationRateLimit = v
	}

	if v, err := c.GetConfigInt("RISK_MAX_SCORE"); err != nil {
		errs = append(errs, fmt.Sprintf("RISK_MAX_SCORE: %v", err))
	} else {
		cfg.RiskMaxScore = v
	}

	if v, err := c.GetConfigFloat("KYB_FACE_SCORE_THRESHOLD"); err != nil {
		errs = append(errs, fmt.Sprintf("KYB_FACE_SCORE_THRESHOLD: %v", err))
	} else {
		cfg.KYBFaceScoreThreshold = v
	}

	if v, err := c.GetConfigDuration("BLACKLIST_CACHE_TTL"); err != nil {
		errs = append(errs, fmt.Sprintf("BLACKLIST_CACHE_TTL: %v", err))
	} else {
		cfg.BlacklistCacheTTL = v
	}

	if v, err := c.GetConfigInt("SLIDING_WINDOW_REG_LIMIT"); err != nil {
		errs = append(errs, fmt.Sprintf("SLIDING_WINDOW_REG_LIMIT: %v", err))
	} else {
		cfg.SlidingWindowRegLimit = v
	}

	if v, err := c.GetConfigDuration("SLIDING_WINDOW_REG_WINDOW"); err != nil {
		errs = append(errs, fmt.Sprintf("SLIDING_WINDOW_REG_WINDOW: %v", err))
	} else {
		cfg.SlidingWindowRegWindow = v
	}

	if v, err := c.GetConfigInt("SLIDING_WINDOW_REF_ABUSE_LIMIT"); err != nil {
		errs = append(errs, fmt.Sprintf("SLIDING_WINDOW_REF_ABUSE_LIMIT: %v", err))
	} else {
		cfg.SlidingWindowRefAbuseLimit = v
	}

	if v, err := c.GetConfigDuration("SLIDING_WINDOW_REF_ABUSE_WINDOW"); err != nil {
		errs = append(errs, fmt.Sprintf("SLIDING_WINDOW_REF_ABUSE_WINDOW: %v", err))
	} else {
		cfg.SlidingWindowRefAbuseWindow = v
	}

	if v, err := c.GetConfigInt("AUDIT_LOG_DEFAULT_PAGE_SIZE"); err != nil {
		errs = append(errs, fmt.Sprintf("AUDIT_LOG_DEFAULT_PAGE_SIZE: %v", err))
	} else {
		cfg.AuditLogDefaultLimit = v
	}

	if v, err := c.GetConfigInt("RISK_HISTORY_DEFAULT_LIMIT"); err != nil {
		errs = append(errs, fmt.Sprintf("RISK_HISTORY_DEFAULT_LIMIT: %v", err))
	} else {
		cfg.RiskHistoryDefaultLimit = v
	}

	cfg.EncryptionKey = ""

	if len(errs) > 0 {
		return nil, fmt.Errorf("config load failed:\n  %s", strings.Join(errs, "\n  "))
	}

	return cfg, nil
}
