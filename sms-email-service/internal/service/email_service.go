package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/trigold786/92-Account-Center/sms-email-service/internal/provider"

	"github.com/redis/go-redis/v9"
)

const (
	emailCodeLength  = 6
	emailOTPKey      = "otp:email:"
	emailRateLimit   = "rate_limit:email:"
	emailDailyLimit  = 10
	emailRateLimitTTL = 2 * time.Minute
	emailOTPTTL      = 5 * time.Minute
)

var ErrEmailRateLimit = errors.New("email rate limit exceeded")

type EmailService interface {
	SendVerificationCode(ctx context.Context, email string) error
	VerifyCode(ctx context.Context, email, code string) (bool, error)
}

type emailService struct {
	redis  *redis.Client
	provider provider.EmailProvider
}

func NewEmailService(p provider.EmailProvider) EmailService {
	return &emailService{
		redis:    globalRedis,
		provider: p,
	}
}

func (s *emailService) SendVerificationCode(ctx context.Context, email string) error {
	rlKey := emailRateLimit + email
	exists, err := s.redis.Exists(ctx, rlKey).Result()
	if err == nil && exists > 0 {
		return ErrEmailRateLimit
	}

	dailyKey := "rate_limit:email:daily:" + email
	dailyCount, err := s.redis.Get(ctx, dailyKey).Int()
	if err == nil && dailyCount >= emailDailyLimit {
		return ErrEmailRateLimit
	}

	code := generateEmailCode(emailCodeLength)

	if err := s.provider.SendVerificationCode(ctx, email, code); err != nil {
		return err
	}

	s.redis.Set(ctx, emailOTPKey+email, code, emailOTPTTL)
	s.redis.Set(ctx, rlKey, "1", emailRateLimitTTL)
	s.redis.Incr(ctx, dailyKey)
	s.redis.Expire(ctx, dailyKey, 24*time.Hour)

	return nil
}

func (s *emailService) VerifyCode(ctx context.Context, email, code string) (bool, error) {
	if code == "" || len(code) != emailCodeLength {
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
