package svcconfig

import (
	"fmt"
	"strings"

	"github.com/trigold786/92-Account-Center/pkg/config"
)

type DataProductConfig struct {
	DashboardTrendDays int
}

func Load(c *config.Client) (*DataProductConfig, error) {
	cfg := &DataProductConfig{
		DashboardTrendDays: 30,
	}
	var errs []string

	if v, err := c.GetConfigInt("DASHBOARD_TREND_DAYS"); err != nil {
		errs = append(errs, fmt.Sprintf("DASHBOARD_TREND_DAYS: %v", err))
	} else {
		cfg.DashboardTrendDays = v
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("config load failed:\n  %s", strings.Join(errs, "\n  "))
	}

	return cfg, nil
}
