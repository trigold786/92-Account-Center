package svcconfig

import (
	"fmt"
	"time"

	"github.com/trigold786/92-Account-Center/pkg/config"
)

type AccountConfig struct {
	DeletionFreezeDays          int
	SubscriptionDefaultDuration time.Duration
	EntitlementCacheTTL         time.Duration
}

func Load(c *config.Client) (*AccountConfig, error) {
	cfg := &AccountConfig{}
	var errs []string

	if v, err := c.GetConfigInt("ACCOUNT_DELETION_FREEZE_DAYS"); err != nil {
		errs = append(errs, fmt.Sprintf("ACCOUNT_DELETION_FREEZE_DAYS: %v", err))
	} else {
		cfg.DeletionFreezeDays = v
	}

	if v, err := c.GetConfigDuration("SUBSCRIPTION_DEFAULT_DURATION"); err != nil {
		errs = append(errs, fmt.Sprintf("SUBSCRIPTION_DEFAULT_DURATION: %v", err))
	} else {
		cfg.SubscriptionDefaultDuration = v
	}

	if v, err := c.GetConfigDuration("ENTITLEMENT_CACHE_TTL"); err != nil {
		errs = append(errs, fmt.Sprintf("ENTITLEMENT_CACHE_TTL: %v", err))
	} else {
		cfg.EntitlementCacheTTL = v
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
