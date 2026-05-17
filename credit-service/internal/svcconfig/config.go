package svcconfig

import (
	"fmt"
	"strings"
	"time"

	"github.com/trigold786/92-Account-Center/pkg/config"
)

type CreditConfig struct {
	DefaultPageSize           int
	DefaultRebateRate         float64
	ReferralLinkTemplate      string
	SubscriptionStreamKey     string
	SubscriptionConsumerGroup string
	SubscriptionConsumerID    string
	WorkerPollInterval        time.Duration
	WorkerBatchSize           int
}

func Load(c *config.Client) (*CreditConfig, error) {
	cfg := &CreditConfig{
		DefaultPageSize:           20,
		DefaultRebateRate:         0.10,
		ReferralLinkTemplate:      "https://app.example.com/referral?code=%s",
		SubscriptionStreamKey:     "subscription:paid",
		SubscriptionConsumerGroup: "credit-rebate-group",
		SubscriptionConsumerID:    "credit-worker-1",
		WorkerPollInterval:        2 * time.Second,
		WorkerBatchSize:           10,
	}
	var errs []string

	if v, err := c.GetConfigInt("CREDIT_PAGE_SIZE"); err != nil {
		errs = append(errs, fmt.Sprintf("CREDIT_PAGE_SIZE: %v", err))
	} else {
		cfg.DefaultPageSize = v
	}

	if v, err := c.GetConfigFloat("CREDIT_DEFAULT_REBATE_RATE"); err != nil {
		errs = append(errs, fmt.Sprintf("CREDIT_DEFAULT_REBATE_RATE: %v", err))
	} else {
		cfg.DefaultRebateRate = v
	}

	if v, err := c.GetConfig("CREDIT_REFERRAL_LINK_TEMPLATE"); err != nil {
		errs = append(errs, fmt.Sprintf("CREDIT_REFERRAL_LINK_TEMPLATE: %v", err))
	} else {
		cfg.ReferralLinkTemplate = v
	}

	if v, err := c.GetConfig("CREDIT_SUBSCRIPTION_STREAM_KEY"); err != nil {
		errs = append(errs, fmt.Sprintf("CREDIT_SUBSCRIPTION_STREAM_KEY: %v", err))
	} else {
		cfg.SubscriptionStreamKey = v
	}

	if v, err := c.GetConfig("CREDIT_SUBSCRIPTION_CONSUMER_GROUP"); err != nil {
		errs = append(errs, fmt.Sprintf("CREDIT_SUBSCRIPTION_CONSUMER_GROUP: %v", err))
	} else {
		cfg.SubscriptionConsumerGroup = v
	}

	if v, err := c.GetConfig("CREDIT_SUBSCRIPTION_CONSUMER_ID"); err != nil {
		errs = append(errs, fmt.Sprintf("CREDIT_SUBSCRIPTION_CONSUMER_ID: %v", err))
	} else {
		cfg.SubscriptionConsumerID = v
	}

	if v, err := c.GetConfigDuration("CREDIT_WORKER_POLL_INTERVAL"); err != nil {
		errs = append(errs, fmt.Sprintf("CREDIT_WORKER_POLL_INTERVAL: %v", err))
	} else {
		cfg.WorkerPollInterval = v
	}

	if v, err := c.GetConfigInt("CREDIT_WORKER_BATCH_SIZE"); err != nil {
		errs = append(errs, fmt.Sprintf("CREDIT_WORKER_BATCH_SIZE: %v", err))
	} else {
		cfg.WorkerBatchSize = v
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("config load failed:\n  %s", strings.Join(errs, "\n  "))
	}

	return cfg, nil
}
