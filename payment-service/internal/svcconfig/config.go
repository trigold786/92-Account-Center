package svcconfig

import (
	"fmt"
	"strings"

	"github.com/trigold786/92-Account-Center/pkg/config"
)

type PaymentConfig struct {
	DefaultPageSize    int
	OrderExpiryMinutes int
}

func Load(c *config.Client) (*PaymentConfig, error) {
	cfg := &PaymentConfig{
		DefaultPageSize:    20,
		OrderExpiryMinutes: 30,
	}
	var errs []string

	if v, err := c.GetConfigInt("PAYMENT_PAGE_SIZE"); err != nil {
		errs = append(errs, fmt.Sprintf("PAYMENT_PAGE_SIZE: %v", err))
	} else {
		cfg.DefaultPageSize = v
	}

	if v, err := c.GetConfigInt("PAYMENT_ORDER_EXPIRY_MINUTES"); err != nil {
		errs = append(errs, fmt.Sprintf("PAYMENT_ORDER_EXPIRY_MINUTES: %v", err))
	} else {
		cfg.OrderExpiryMinutes = v
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("config load failed:\n  %s", strings.Join(errs, "\n  "))
	}

	return cfg, nil
}
