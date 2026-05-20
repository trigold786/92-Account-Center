package svcconfig

import (
	"os"
	"strconv"
	"time"

	"github.com/trigold786/92-Account-Center/pkg/config"
)

type GatewayConfig struct {
	AccountServiceURL      string
	AuthServiceURL         string
	NotificationServiceURL string
	CreditServiceURL       string
	ComplianceServiceURL   string
	DataProductServiceURL  string
	JWTSecret              string
	Port                   string
	RateLimitRPS           int
	CacheMaxAge            int
	ShutdownTimeout          time.Duration
	MaxDesensitizeBodySize   int64
	ResponseHeaderTimeoutSec int
	IdleConnTimeoutSec       int
	GlobalRequestTimeoutSec  int
}

func Load(c *config.Client) (*GatewayConfig, error) {
	cfg := &GatewayConfig{}

	cfg.AccountServiceURL = loadString(c, "ACCOUNT_SERVICE_URL", "ACCOUNT_SERVICE_URL", "http://localhost:30301")
	cfg.AuthServiceURL = loadString(c, "AUTH_SERVICE_URL", "AUTH_SERVICE_URL", "http://localhost:30302")
	cfg.NotificationServiceURL = loadString(c, "NOTIFICATION_SERVICE_URL", "NOTIFICATION_SERVICE_URL", "http://localhost:30311")
	cfg.CreditServiceURL = loadString(c, "CREDIT_SERVICE_URL", "CREDIT_SERVICE_URL", "http://localhost:30312")
	cfg.ComplianceServiceURL = loadString(c, "COMPLIANCE_SERVICE_URL", "COMPLIANCE_SERVICE_URL", "http://localhost:30313")
	cfg.DataProductServiceURL = loadString(c, "DATA_PRODUCT_SERVICE_URL", "DATA_PRODUCT_SERVICE_URL", "http://localhost:30314")
	cfg.JWTSecret = loadString(c, "JWT_SECRET", "JWT_SECRET", "default-secret")
	cfg.Port = loadString(c, "GATEWAY_PORT", "PORT", "30300")
	cfg.RateLimitRPS = loadInt(c, "GATEWAY_RATE_LIMIT_RPS", "RATE_LIMIT_RPS", 100)
	cfg.CacheMaxAge = loadInt(c, "GATEWAY_CACHE_MAX_AGE", "CACHE_MAX_AGE", 60)
	cfg.ShutdownTimeout = time.Duration(loadInt(c, "GATEWAY_SHUTDOWN_TIMEOUT_SECONDS", "SHUTDOWN_TIMEOUT_SECONDS", 10)) * time.Second
	cfg.MaxDesensitizeBodySize = int64(loadInt(c, "GATEWAY_MAX_DESENSITIZE_BODY_SIZE", "MAX_DESENSITIZE_BODY_SIZE", 1048576))
	cfg.ResponseHeaderTimeoutSec = loadInt(c, "GATEWAY_RESPONSE_HEADER_TIMEOUT_SEC", "RESPONSE_HEADER_TIMEOUT_SEC", 30)
	cfg.IdleConnTimeoutSec = loadInt(c, "GATEWAY_IDLE_CONN_TIMEOUT_SEC", "IDLE_CONN_TIMEOUT_SEC", 90)
	cfg.GlobalRequestTimeoutSec = loadInt(c, "GATEWAY_GLOBAL_REQUEST_TIMEOUT_SEC", "GLOBAL_REQUEST_TIMEOUT_SEC", 60)

	return cfg, nil
}

func loadString(c *config.Client, configCode, envKey, defaultVal string) string {
	if v, err := c.GetConfig(configCode); err == nil && v != "" {
		return v
	}
	if v := os.Getenv(envKey); v != "" {
		return v
	}
	return defaultVal
}

func loadInt(c *config.Client, configCode, envKey string, defaultVal int) int {
	if v, err := c.GetConfigInt(configCode); err == nil {
		return v
	}
	if v := os.Getenv(envKey); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
