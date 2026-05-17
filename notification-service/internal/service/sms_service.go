package service

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/notification-service/internal/provider"
	"github.com/trigold786/92-Account-Center/notification-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/notification-service/pkg/circuitbreaker"
)

var (
	ErrSMSRateLimit       = errors.New("rate limit exceeded")
	ErrNoAvailableProvider = errors.New("no available SMS provider")
)

const (
	rateLimitKey    = "rate_limit:sms:"
	dailyKeyPrefix  = "rate_limit:sms:daily:"
	smsOTPKeyPrefix = "otp:"
)

type SMSService interface {
	SendCode(ctx context.Context, phoneNumber string) error
	VerifyCode(ctx context.Context, phoneNumber, code string) (bool, error)
	GetProviderStatus() []ProviderStatus
}

type smsService struct {
	redis           *redis.Client
	providers       []provider.SMSProvider
	circuitBreakers []*circuitbreaker.CircuitBreaker
	currentIndex    int64
	cfg             *svcconfig.NotificationConfig
}

type ProviderStatus struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

var globalRedis *redis.Client

func SetRedisClient(client *redis.Client) {
	globalRedis = client
}

func NewSMSService(providers []provider.SMSProvider, circuitBreakers []*circuitbreaker.CircuitBreaker, cfg *svcconfig.NotificationConfig) SMSService {
	return &smsService{
		redis:           globalRedis,
		providers:       providers,
		circuitBreakers: circuitBreakers,
		currentIndex:    0,
		cfg:             cfg,
	}
}

func (s *smsService) SendCode(ctx context.Context, phoneNumber string) error {
	rlKey := rateLimitKey + phoneNumber
	exists, err := s.redis.Exists(ctx, rlKey).Result()
	if err == nil && exists > 0 {
		return ErrSMSRateLimit
	}

	dailyKey := dailyKeyPrefix + phoneNumber
	dailyCount, err := s.redis.Get(ctx, dailyKey).Int()
	if err == nil && dailyCount >= s.cfg.SMSDailyLimit {
		return ErrSMSRateLimit
	}

	var lastErr error
	for i := 0; i < len(s.providers); i++ {
		idx := (int(atomic.LoadInt64(&s.currentIndex)) + i) % len(s.providers)
		cb := s.circuitBreakers[idx]

		var code string
		err := cb.Execute(func() error {
			var sendErr error
			code, sendErr = s.providers[idx].SendCode(ctx, phoneNumber)
			return sendErr
		})

		if err != nil {
			lastErr = err
			continue
		}

		atomic.StoreInt64(&s.currentIndex, int64(idx))

		s.redis.Set(ctx, smsOTPKeyPrefix+phoneNumber, code, s.cfg.SMSOTPTTL)
		s.redis.Set(ctx, rlKey, "1", s.cfg.SMSRateLimitTTL)
		s.redis.Incr(ctx, dailyKey)
		s.redis.Expire(ctx, dailyKey, 24*time.Hour)

		return nil
	}

	return lastErr
}

func (s *smsService) VerifyCode(ctx context.Context, phoneNumber, code string) (bool, error) {
	if code == "" || len(code) != 6 {
		return false, nil
	}
	stored, err := s.redis.Get(ctx, smsOTPKeyPrefix+phoneNumber).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return stored == code, nil
}

func (s *smsService) GetProviderStatus() []ProviderStatus {
	statuses := make([]ProviderStatus, len(s.providers))
	for i, p := range s.providers {
		state := s.circuitBreakers[i].State()
		statusStr := "closed"
		switch state {
		case circuitbreaker.StateOpen:
			statusStr = "open"
		case circuitbreaker.StateHalfOpen:
			statusStr = "half-open"
		}
		statuses[i] = ProviderStatus{Name: p.Name(), Status: statusStr}
	}
	return statuses
}
