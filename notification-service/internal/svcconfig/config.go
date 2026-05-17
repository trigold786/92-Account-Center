package svcconfig

import (
	"fmt"
	"time"

	"github.com/trigold786/92-Account-Center/pkg/config"
)

type NotificationConfig struct {
	SMSRateLimitTTL    time.Duration
	SMSDailyLimit      int
	SMSOTPTTL          time.Duration
	SMSCodeLength      int
	EmailCodeLength    int
	EmailDailyLimit    int
	EmailRateLimitTTL  time.Duration
	EmailOTPTTL        time.Duration
	OTPLength          int
	MagicLinkTTL       time.Duration
	RateLimitMax       int
	RateLimitWindow    time.Duration
}

func Load(c *config.Client) (*NotificationConfig, error) {
	cfg := &NotificationConfig{
		SMSRateLimitTTL:   2 * time.Minute,
		SMSDailyLimit:     10,
		SMSOTPTTL:         5 * time.Minute,
		SMSCodeLength:     6,
		EmailCodeLength:   6,
		EmailDailyLimit:   10,
		EmailRateLimitTTL: 2 * time.Minute,
		EmailOTPTTL:       5 * time.Minute,
		OTPLength:         6,
		MagicLinkTTL:      15 * time.Minute,
		RateLimitMax:      3,
		RateLimitWindow:   time.Hour,
	}
	var errs []string

	if v, err := c.GetConfigDuration("SMS_CODE_EXPIRE"); err != nil {
		errs = append(errs, fmt.Sprintf("SMS_CODE_EXPIRE: %v", err))
	} else {
		cfg.SMSOTPTTL = v
	}
	if v, err := c.GetConfigInt("SMS_CODE_LENGTH"); err != nil {
		errs = append(errs, fmt.Sprintf("SMS_CODE_LENGTH: %v", err))
	} else {
		cfg.SMSCodeLength = v
	}
	if v, err := c.GetConfigInt("SMS_DAILY_LIMIT"); err != nil {
		errs = append(errs, fmt.Sprintf("SMS_DAILY_LIMIT: %v", err))
	} else {
		cfg.SMSDailyLimit = v
	}
	if v, err := c.GetConfigDuration("SMS_RATE_LIMIT_PER_USER"); err != nil {
		errs = append(errs, fmt.Sprintf("SMS_RATE_LIMIT_PER_USER: %v", err))
	} else {
		cfg.SMSRateLimitTTL = v
	}
	if v, err := c.GetConfigDuration("EMAIL_OTP_EXPIRE"); err != nil {
		errs = append(errs, fmt.Sprintf("EMAIL_OTP_EXPIRE: %v", err))
	} else {
		cfg.EmailOTPTTL = v
	}
	if v, err := c.GetConfigDuration("EMAIL_MAGIC_LINK_EXPIRE"); err != nil {
		errs = append(errs, fmt.Sprintf("EMAIL_MAGIC_LINK_EXPIRE: %v", err))
	} else {
		cfg.MagicLinkTTL = v
	}
	if v, err := c.GetConfigDuration("EMAIL_RATE_LIMIT_PER_USER"); err != nil {
		errs = append(errs, fmt.Sprintf("EMAIL_RATE_LIMIT_PER_USER: %v", err))
	} else {
		cfg.EmailRateLimitTTL = v
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("config load failed:\n  %s", join(errs, "\n  "))
	}

	return cfg, nil
}

func join(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}
