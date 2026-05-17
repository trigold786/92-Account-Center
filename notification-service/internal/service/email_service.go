package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/notification-service/internal/provider"
	"github.com/trigold786/92-Account-Center/notification-service/internal/svcconfig"
)

const (
	emailOTPKey    = "otp:email:"
	emailRateLimit = "rate_limit:email:"
)

var ErrEmailRateLimit = errors.New("email rate limit exceeded")

type SimpleEmailService interface {
	SendVerificationCode(ctx context.Context, email string) error
	VerifyCode(ctx context.Context, email, code string) (bool, error)
}

type simpleEmailService struct {
	redis    *redis.Client
	provider provider.VerificationEmailProvider
	cfg      *svcconfig.NotificationConfig
}

func NewSimpleEmailService(p provider.VerificationEmailProvider, cfg *svcconfig.NotificationConfig) SimpleEmailService {
	return &simpleEmailService{
		redis:    globalRedis,
		provider: p,
		cfg:      cfg,
	}
}

func (s *simpleEmailService) SendVerificationCode(ctx context.Context, email string) error {
	rlKey := emailRateLimit + email
	exists, err := s.redis.Exists(ctx, rlKey).Result()
	if err == nil && exists > 0 {
		return ErrEmailRateLimit
	}

	dailyKey := "rate_limit:email:daily:" + email
	dailyCount, err := s.redis.Get(ctx, dailyKey).Int()
	if err == nil && dailyCount >= s.cfg.EmailDailyLimit {
		return ErrEmailRateLimit
	}

	code := generateEmailCode(s.cfg.EmailCodeLength)

	if err := s.provider.SendVerificationCode(ctx, email, code); err != nil {
		return err
	}

	s.redis.Set(ctx, emailOTPKey+email, code, s.cfg.EmailOTPTTL)
	s.redis.Set(ctx, rlKey, "1", s.cfg.EmailRateLimitTTL)
	s.redis.Incr(ctx, dailyKey)
	s.redis.Expire(ctx, dailyKey, 24*time.Hour)

	return nil
}

func (s *simpleEmailService) VerifyCode(ctx context.Context, email, code string) (bool, error) {
	if code == "" || len(code) != s.cfg.EmailCodeLength {
		return false, nil
	}
	stored, err := s.redis.Get(ctx, emailOTPKey+email).Result()
	if err != nil {
		if err == redis.Nil {
			return false, nil
		}
		return false, err
	}
	return stored == code, nil
}

func generateEmailCode(length int) string {
	const digits = "0123456789"
	result := make([]byte, length)
	for i := range result {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		result[i] = digits[n.Int64()]
	}
	return string(result)
}
