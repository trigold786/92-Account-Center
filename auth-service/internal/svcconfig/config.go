package svcconfig

import (
	"fmt"
	"strings"
	"time"

	"github.com/trigold786/92-Account-Center/pkg/config"
)

type AuthConfig struct {
	JwtSecretKey         string `json:"-"`
	JwtAccessTokenExpire time.Duration
	JwtRefreshTokenExpire time.Duration
	LoginMaxAttempts     int
	LoginLockoutDuration time.Duration
	SessionMaxPerUser    int
	SessionTimeout       time.Duration
	SessionSlidingWindowEnabled bool
	SessionRenewalAdvanceTime   time.Duration
	DeviceTrustDays      int
	DeviceRiskThreshold  float64
	QRCodeExpire         time.Duration
	LoginRateLimitPerIP  int
}

func Load(c *config.Client) (*AuthConfig, error) {
	cfg := &AuthConfig{}

	var errs []string

	if v, err := c.GetConfigDuration("JWT_ACCESS_TOKEN_EXPIRE"); err != nil {
		errs = append(errs, fmt.Sprintf("JWT_ACCESS_TOKEN_EXPIRE: %v", err))
	} else {
		cfg.JwtAccessTokenExpire = v
	}

	if v, err := c.GetConfigDuration("JWT_REFRESH_TOKEN_EXPIRE"); err != nil {
		errs = append(errs, fmt.Sprintf("JWT_REFRESH_TOKEN_EXPIRE: %v", err))
	} else {
		cfg.JwtRefreshTokenExpire = v
	}

	if v, err := c.GetConfigInt("LOGIN_MAX_ATTEMPTS"); err != nil {
		errs = append(errs, fmt.Sprintf("LOGIN_MAX_ATTEMPTS: %v", err))
	} else {
		cfg.LoginMaxAttempts = v
	}

	if v, err := c.GetConfigDuration("LOGIN_LOCKOUT_DURATION"); err != nil {
		errs = append(errs, fmt.Sprintf("LOGIN_LOCKOUT_DURATION: %v", err))
	} else {
		cfg.LoginLockoutDuration = v
	}

	if v, err := c.GetConfigInt("SESSION_MAX_PER_USER"); err != nil {
		errs = append(errs, fmt.Sprintf("SESSION_MAX_PER_USER: %v", err))
	} else {
		cfg.SessionMaxPerUser = v
	}

	if v, err := c.GetConfigDuration("SESSION_IDLE_TIMEOUT"); err != nil {
		errs = append(errs, fmt.Sprintf("SESSION_IDLE_TIMEOUT: %v", err))
	} else {
		cfg.SessionTimeout = v
	}

	if v, err := c.GetConfigBool("SESSION_SLIDING_WINDOW_ENABLED"); err != nil {
		errs = append(errs, fmt.Sprintf("SESSION_SLIDING_WINDOW_ENABLED: %v", err))
	} else {
		cfg.SessionSlidingWindowEnabled = v
	}

	if v, err := c.GetConfigDuration("SESSION_RENEWAL_ADVANCE_TIME"); err != nil {
		errs = append(errs, fmt.Sprintf("SESSION_RENEWAL_ADVANCE_TIME: %v", err))
	} else {
		cfg.SessionRenewalAdvanceTime = v
	}

	if v, err := c.GetConfigInt("DEVICE_DEFAULT_TRUST_DAYS"); err != nil {
		errs = append(errs, fmt.Sprintf("DEVICE_DEFAULT_TRUST_DAYS: %v", err))
	} else {
		cfg.DeviceTrustDays = v
	}

	if v, err := c.GetConfigDuration("QR_CODE_EXPIRE"); err != nil {
		errs = append(errs, fmt.Sprintf("QR_CODE_EXPIRE: %v", err))
	} else {
		cfg.QRCodeExpire = v
	}

	if v, err := c.GetConfigInt("LOGIN_RATE_LIMIT_PER_IP"); err != nil {
		errs = append(errs, fmt.Sprintf("LOGIN_RATE_LIMIT_PER_IP: %v", err))
	} else {
		cfg.LoginRateLimitPerIP = v
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("config load failed:\n  %s", strings.Join(errs, "\n  "))
	}

	cfg.DeviceRiskThreshold = 0.3

	return cfg, nil
}
