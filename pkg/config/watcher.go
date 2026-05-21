package config

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

type Config struct {
	Values map[string]string
}

type ConfigWatcher struct {
	client   *Client
	interval time.Duration
	store    atomic.Value
	logger   *slog.Logger
}

func NewWatcher(client *Client, interval time.Duration) *ConfigWatcher {
	return &ConfigWatcher{
		client:   client,
		interval: interval,
		logger:   slog.Default(),
	}
}

func (w *ConfigWatcher) Watch(ctx context.Context, onChange func(*Config)) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.poll(ctx, onChange)
		}
	}
}

func (w *ConfigWatcher) poll(ctx context.Context, onChange func(*Config)) {
	codes := []string{
		"JWT_SECRET", "GATEWAY_RATE_LIMIT_RPS", "GATEWAY_CACHE_MAX_AGE",
		"JWT_ACCESS_TOKEN_EXPIRE", "JWT_REFRESH_TOKEN_EXPIRE",
		"LOGIN_MAX_ATTEMPTS", "LOGIN_LOCKOUT_DURATION",
		"SESSION_TIMEOUT", "SESSION_MAX_PER_USER",
	}

	newValues := make(map[string]string)
	for _, code := range codes {
		val, err := w.client.GetConfig(code)
		if err != nil {
			continue
		}
		newValues[code] = val
	}

	if len(newValues) == 0 {
		return
	}

	newCfg := &Config{Values: newValues}
	w.store.Store(newCfg)

	if onChange != nil {
		onChange(newCfg)
	}
}

func (w *ConfigWatcher) Get() *Config {
	val := w.store.Load()
	if val == nil {
		return nil
	}
	return val.(*Config)
}
