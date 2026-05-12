package service

import (
	"context"
	"errors"
	"sync"
	"time"

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

type loginAttempt struct {
	count       int
	lastAttempt time.Time
}

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
}

type authService struct {
	userRepo      UserRepository
	jwtMgr        *jwt.JWTManager
	loginAttempts map[string]*loginAttempt
	attemptMu     sync.Mutex
	blacklist     map[string]time.Time
	blacklistMu   sync.RWMutex
}

func NewAuthService(userRepo UserRepository, jwtMgr *jwt.JWTManager) AuthService {
	return &authService{
		userRepo:      userRepo,
		jwtMgr:        jwtMgr,
		loginAttempts: make(map[string]*loginAttempt),
		blacklist:     make(map[string]time.Time),
	}
}

func (s *authService) isLocked(credential string) bool {
	s.attemptMu.Lock()
	defer s.attemptMu.Unlock()

	attempt, exists := s.loginAttempts[credential]
	if !exists {
		return false
	}
	if attempt.count >= maxFailedAttempts && time.Since(attempt.lastAttempt) < lockoutDuration {
		return true
	}
	if attempt.count >= maxFailedAttempts && time.Since(attempt.lastAttempt) >= lockoutDuration {
		delete(s.loginAttempts, credential)
	}
	return false
}

func (s *authService) recordFailedAttempt(credential string) {
	s.attemptMu.Lock()
	defer s.attemptMu.Unlock()

	attempt, exists := s.loginAttempts[credential]
	if !exists {
		attempt = &loginAttempt{}
		s.loginAttempts[credential] = attempt
	}
	attempt.count++
	attempt.lastAttempt = time.Now()
}

func (s *authService) resetFailedAttempts(credential string) {
	s.attemptMu.Lock()
	defer s.attemptMu.Unlock()
	delete(s.loginAttempts, credential)
}

func (s *authService) isTokenBlacklisted(tokenString string) bool {
	s.blacklistMu.RLock()
	defer s.blacklistMu.RUnlock()

	expiry, exists := s.blacklist[tokenString]
	if !exists {
		return false
	}
	if time.Now().After(expiry) {
		return false
	}
	return true
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

	accessToken, refreshToken, err := s.jwtMgr.GenerateTokenPair(user.ID, user.AccountID)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(24 * time.Hour.Seconds()),
		UserID:       user.ID,
		AccountID:    user.AccountID,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, error) {
	if s.isTokenBlacklisted(refreshToken) {
		return nil, errors.New("token has been revoked")
	}

	accessToken, err := s.jwtMgr.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, err
	}

	claims, err := s.jwtMgr.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(24 * time.Hour.Seconds()),
		UserID:       claims.UserID,
		AccountID:    claims.AccountID,
	}, nil
}

func (s *authService) Logout(ctx context.Context, accessToken string) error {
	claims, err := s.jwtMgr.ValidateToken(accessToken)
	if err != nil {
		return errors.New("invalid token")
	}

	s.blacklistMu.Lock()
	defer s.blacklistMu.Unlock()
	s.blacklist[accessToken] = claims.ExpiresAt.Time

	return nil
}

func init() {
	go func() {
		for {
			time.Sleep(1 * time.Hour)
		}
	}()
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
	if req.Code == "" {
		return false
	}
	return VerifyTOTPCode(user.MFASecret, req.Code)
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
