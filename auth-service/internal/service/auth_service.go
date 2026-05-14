package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
	"github.com/trigold786/92-Account-Center/auth-service/internal/util"
	"github.com/trigold786/92-Account-Center/auth-service/pkg/crypto"
	"github.com/trigold786/92-Account-Center/auth-service/pkg/jwt"
)

var ErrAccountLocked = errors.New("account is temporarily locked due to too many failed login attempts")

const (
	maxFailedAttempts = 5
	lockoutDuration   = 30 * time.Minute
)

type UserRepository interface {
	GetByPhoneNumber(ctx context.Context, phoneNumber string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByAccountID(ctx context.Context, accountID string) (*model.User, error)
}

type AuthService interface {
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, error)
	VerifyPassword(password, storedHash string) bool
	Logout(ctx context.Context, accessToken string) error
	RegisterBiometric(ctx context.Context, req *model.BiometricRegisterRequest) error
	LoginWithBiometric(ctx context.Context, req *model.BiometricLoginRequest) (*model.LoginResponse, error)
}

type authService struct {
	userRepo UserRepository
	jwtMgr   *jwt.JWTManager
	rdb      *redis.Client
}

func NewAuthService(userRepo UserRepository, jwtMgr *jwt.JWTManager, rdb *redis.Client) AuthService {
	return &authService{
		userRepo: userRepo,
		jwtMgr:   jwtMgr,
		rdb:      rdb,
	}
}

func (s *authService) isLocked(credential string) bool {
	if s.rdb == nil {
		return false
	}
	key := "lockout:" + credential
	val, err := s.rdb.Get(context.Background(), key).Result()
	if err != nil {
		return false
	}
	return val == "1"
}

func (s *authService) recordFailedAttempt(credential string) {
	if s.rdb == nil {
		return
	}
	key := "lockout:" + credential
	countKey := "lockout:count:" + credential

	count, err := s.rdb.Incr(context.Background(), countKey).Result()
	if err != nil {
		return
	}
	if count == 1 {
		s.rdb.Expire(context.Background(), countKey, lockoutDuration)
	}

	if count >= maxFailedAttempts {
		s.rdb.Set(context.Background(), key, "1", lockoutDuration)
	}
}

func (s *authService) resetFailedAttempts(credential string) {
	if s.rdb == nil {
		return
	}
	s.rdb.Del(context.Background(), "lockout:"+credential, "lockout:count:"+credential)
}

func (s *authService) isTokenBlacklisted(tokenString string) bool {
	if s.rdb == nil {
		return false
	}
	key := "blacklist:" + tokenString
	exists, err := s.rdb.Exists(context.Background(), key).Result()
	if err != nil {
		return false
	}
	return exists == 1
}

func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	if req == nil {
		return nil, errors.New("login request cannot be nil")
	}

	if s.isLocked(req.Credential) {
		return nil, ErrAccountLocked
	}

	credType := util.IdentifyCredentialType(req.Credential)
	var user *model.User
	var err error

	switch credType {
	case util.CredentialTypePhone:
		user, err = s.userRepo.GetByPhoneNumber(ctx, req.Credential)
	case util.CredentialTypeEmail:
		user, err = s.userRepo.GetByEmail(ctx, req.Credential)
	case util.CredentialTypeAccountID:
		user, err = s.userRepo.GetByAccountID(ctx, req.Credential)
	default:
		return nil, errors.New("invalid credential format")
	}

	if err != nil {
		return nil, err
	}
	if user == nil {
		s.recordFailedAttempt(req.Credential)
		return nil, errors.New("invalid credentials")
	}

	switch {
	case req.Password != "":
		if !s.VerifyPassword(req.Password, user.PasswordHash) {
			s.recordFailedAttempt(req.Credential)
			return nil, errors.New("invalid credentials")
		}
	case req.Code != "":
		if !s.verifyOTPCode(ctx, req.Credential, req.Code) {
			s.recordFailedAttempt(req.Credential)
			return nil, errors.New("invalid or expired verification code")
		}
	default:
		return nil, errors.New("password or verification code is required")
	}

	if user.MFAEnabled && !s.verifyMFA(ctx, user, req) {
		return nil, errors.New("MFA verification required")
	}

	s.resetFailedAttempts(req.Credential)

	var tokenResp *jwt.TokenResponse
	if req.DeviceFingerprintID != "" {
		tokenResp, err = s.jwtMgr.GenerateTokenPairWithDevice(user.ID, user.AccountID, req.DeviceFingerprintID)
	} else {
		tokenResp, err = s.jwtMgr.GenerateTokenPairWithDevice(user.ID, user.AccountID, "")
	}
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:       tokenResp.AccessToken,
		RefreshToken:      tokenResp.RefreshToken,
		ExpiresIn:         tokenResp.ExpiresIn,
		UserID:            user.ID,
		AccountID:         user.AccountID,
		TokenID:           tokenResp.TokenID,
		DeviceBindingInfo: tokenResp.DeviceBindingInfo,
		IsTrusted:         true,
		IsTrustedDevice:   req.DeviceFingerprintID != "",
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, error) {
	if s.isTokenBlacklisted(refreshToken) {
		return nil, errors.New("token has been revoked")
	}

	tokenResp, err := s.jwtMgr.RefreshTokenPair(refreshToken)
	if err != nil {
		return nil, err
	}

	s.rdb.Set(ctx, "blacklist:"+refreshToken, "1", 30*24*time.Hour)
	claims, err := s.jwtMgr.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:       tokenResp.AccessToken,
		RefreshToken:      tokenResp.RefreshToken,
		ExpiresIn:         tokenResp.ExpiresIn,
		UserID:            claims.UserID,
		AccountID:         claims.AccountID,
		TokenID:           tokenResp.TokenID,
		DeviceBindingInfo: tokenResp.DeviceBindingInfo,
		IsTrusted:         true,
		IsTrustedDevice:   claims.DeviceFingerprint != "",
	}, nil
}

func (s *authService) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.jwtMgr.ValidateToken(accessToken)
	if err != nil {
		return errors.New("invalid token")
	}

	if s.rdb != nil {
		ttl := time.Until(claims.ExpiresAt.Time)
		if ttl > 0 {
			key := "blacklist:" + accessToken
			s.rdb.Set(ctx, key, "1", ttl)
		}
	}

	return nil
}

func (s *authService) VerifyPassword(password, storedHash string) bool {
	parts := splitHash(storedHash)
	if len(parts) != 2 {
		return false
	}
	salt, hash := parts[0], parts[1]
	computed := sm3Hash([]byte(salt + password))
	return subtleCompare(hash, computed)
}

func (s *authService) verifyOTPCode(ctx context.Context, credential, code string) bool {
	if code == "" || len(code) != 6 {
		return false
	}
	return VerifyTOTPCode(credential, code)
}

func (s *authService) verifyMFA(ctx context.Context, user *model.User, req *model.LoginRequest) bool {
	if !user.MFAEnabled {
		return true
	}
	if req.Code == "" || user.MFASecret == nil {
		return false
	}
	return VerifyTOTPCode(*user.MFASecret, req.Code)
}

func sm3Hash(data []byte) string {
	return crypto.SM3Hash(data)
}

func splitHash(stored string) []string {
	for i := len(stored) - 1; i >= 0; i-- {
		if stored[i] == '$' {
			return []string{stored[:i], stored[i+1:]}
		}
	}
	return nil
}

func (s *authService) RegisterBiometric(ctx context.Context, req *model.BiometricRegisterRequest) error {
	key := fmt.Sprintf("biometric:%s:%s", req.UserID, req.DeviceFingerprint)
	hash := crypto.SM3Hash([]byte(req.BiometricToken))
	data, _ := json.Marshal(model.BiometricCredential{
		UserID:           req.UserID,
		DeviceFingerprint: req.DeviceFingerprint,
		BiometricType:    req.BiometricType,
		BiometricTokenHash: hash,
		IsActive:         true,
		CreatedAt:        time.Now(),
	})
	return s.rdb.Set(ctx, key, data, 0).Err()
}

func (s *authService) LoginWithBiometric(ctx context.Context, req *model.BiometricLoginRequest) (*model.LoginResponse, error) {
	pattern := fmt.Sprintf("biometric:*:%s", req.DeviceFingerprint)
	keys, err := s.rdb.Keys(ctx, pattern).Result()
	if err != nil || len(keys) == 0 {
		return nil, errors.New("未找到生物识别凭证")
	}

	for _, key := range keys {
		data, _ := s.rdb.Get(ctx, key).Result()
		var cred model.BiometricCredential
		if err := json.Unmarshal([]byte(data), &cred); err == nil {
			hash := crypto.SM3Hash([]byte(req.BiometricToken))
			if hash == cred.BiometricTokenHash && cred.IsActive {
				var user model.User
				user.ID, _ = strconv.ParseInt(cred.UserID, 10, 64)
				user.AccountID = cred.UserID
				tokenResp, err := s.jwtMgr.GenerateTokenPairWithDevice(user.ID, user.AccountID, req.DeviceFingerprint)
				if err != nil {
					return nil, err
				}
				now := time.Now()
				cred.LastUsedAt = &now
				newData, _ := json.Marshal(cred)
				s.rdb.Set(ctx, key, newData, 0)
				return &model.LoginResponse{
					UserID:            user.ID,
					AccessToken:       tokenResp.AccessToken,
					RefreshToken:      tokenResp.RefreshToken,
					ExpiresIn:         tokenResp.ExpiresIn,
					TokenID:           tokenResp.TokenID,
					DeviceBindingInfo: tokenResp.DeviceBindingInfo,
					IsTrustedDevice:   true,
					IsTrusted:         true,
				}, nil
			}
		}
	}
	return nil, errors.New("生物识别验证失败")
}

func subtleCompare(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var result byte
	for i := 0; i < len(a); i++ {
		result |= a[i] ^ b[i]
	}
	return result == 0
}
