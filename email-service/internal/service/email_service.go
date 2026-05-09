package service

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"time"

	"account-center/email-service/internal/model"
	"account-center/email-service/internal/provider"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

var (
	ErrInvalidOTP       = errors.New("invalid or expired OTP")
	ErrRateLimitExceeded = errors.New("rate limit exceeded: 3 OTP requests per hour")
	ErrInvalidMagicLink = errors.New("invalid or expired magic link")
)

const (
	otpLength           = 6
	otpTTL              = 5 * time.Minute
	magicLinkTTL        = 15 * time.Minute
	rateLimitMax       = 3
	rateLimitWindow    = time.Hour
	otpKeyPrefix       = "otp:"
	rateLimitKeyPrefix = "email_rate:"
)

type EmailService interface {
	SendOTP(ctx context.Context, email string) (*model.OTPResponse, error)
	VerifyOTP(ctx context.Context, email, code string) (bool, error)
	SendMagicLink(ctx context.Context, email, targetURL string) (*model.MagicLinkResponse, error)
	VerifyMagicLink(ctx context.Context, token string) (string, error)
	SendEmail(ctx context.Context, to, subject, content string) error
}

type emailService struct {
	redisClient *redis.Client
	provider    provider.EmailProvider
	jwtSecret   string
	fromAddress string
}

func NewEmailService(redisClient *redis.Client, emailProvider provider.EmailProvider, jwtSecret, fromAddress string) EmailService {
	return &emailService{
		redisClient: redisClient,
		provider:    emailProvider,
		jwtSecret:   jwtSecret,
		fromAddress: fromAddress,
	}
}

func (s *emailService) SendOTP(ctx context.Context, email string) (*model.OTPResponse, error) {
	rateLimitKey := rateLimitKeyPrefix + email

	count, err := s.redisClient.Get(ctx, rateLimitKey).Int()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to check rate limit: %w", err)
	}

	if count >= rateLimitMax {
		return nil, ErrRateLimitExceeded
	}

	otp := generateOTP()

	otpKey := otpKeyPrefix + email
	err = s.redisClient.Set(ctx, otpKey, otp, otpTTL).Err()
	if err != nil {
		return nil, fmt.Errorf("failed to store OTP: %w", err)
	}

	if count == 0 {
		err = s.redisClient.Set(ctx, rateLimitKey, 1, rateLimitWindow).Err()
	} else {
		err = s.redisClient.Incr(ctx, rateLimitKey).Err()
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update rate limit: %w", err)
	}

	subject := "Your Account Center OTP"
	content := fmt.Sprintf("Your one-time password is: <strong>%s</strong>. It expires in 5 minutes.", otp)

	result := s.provider.Send(ctx, email, subject, content)
	if !result.Success {
		return nil, fmt.Errorf("failed to send OTP email: %w", result.Error)
	}

	return &model.OTPResponse{
		ExpiresIn: int(otpTTL.Seconds()),
	}, nil
}

func (s *emailService) VerifyOTP(ctx context.Context, email, code string) (bool, error) {
	otpKey := otpKeyPrefix + email

	storedOTP, err := s.redisClient.Get(ctx, otpKey).String()
	if err != nil {
		if err == redis.Nil {
			return false, ErrInvalidOTP
		}
		return false, fmt.Errorf("failed to get OTP: %w", err)
	}

	if storedOTP != code {
		return false, nil
	}

	s.redisClient.Del(ctx, otpKey)

	return true, nil
}

func (s *emailService) SendMagicLink(ctx context.Context, email, targetURL string) (*model.MagicLinkResponse, error) {
	claims := jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(magicLinkTTL).Unix(),
		"iat":   time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return nil, fmt.Errorf("failed to sign magic link token: %w", err)
	}

	magicLink := fmt.Sprintf("%s?token=%s", targetURL, tokenString)

	subject := "Your Account Center Magic Link"
	content := fmt.Sprintf("Click the link below to sign in:<br><a href=\"%s\">%s</a><br><br>This link expires in 15 minutes.", magicLink, magicLink)

	result := s.provider.Send(ctx, email, subject, content)
	if !result.Success {
		return nil, fmt.Errorf("failed to send magic link email: %w", result.Error)
	}

	return &model.MagicLinkResponse{
		MagicLink: magicLink,
		ExpiresIn: int(magicLinkTTL.Seconds()),
	}, nil
}

func (s *emailService) VerifyMagicLink(ctx context.Context, tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return "", ErrInvalidMagicLink
	}

	if !token.Valid {
		return "", ErrInvalidMagicLink
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", ErrInvalidMagicLink
	}

	email, ok := claims["email"].(string)
	if !ok {
		return "", ErrInvalidMagicLink
	}

	return email, nil
}

func (s *emailService) SendEmail(ctx context.Context, to, subject, content string) error {
	result := s.provider.Send(ctx, to, subject, content)
	if !result.Success {
		return fmt.Errorf("failed to send email: %w", result.Error)
	}
	return nil
}

func generateOTP() string {
	const digits = "0123456789"
	otp := make([]byte, otpLength)
	for i := range otp {
		n, _ := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		otp[i] = digits[n.Int64()]
	}
	return string(otp)
}
