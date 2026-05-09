package service

import (
	"context"
	"errors"
	"time"

	"github.com/sunxi/92-Account-Center/account-service/internal/model"
	"github.com/sunxi/92-Account-Center/account-service/internal/repository"
	"github.com/sunxi/92-Account-Center/auth-service/internal/model"
	"github.com/sunxi/92-Account-Center/auth-service/internal/util"
	"github.com/sunxi/92-Account-Center/auth-service/pkg/jwt"
	"golang.org/x/crypto/bcrypt"

	devicefingerprintservice "github.com/sunxi/92-Account-Center/device-fingerprint-service/internal/service"
)

// AuthService defines the interface for authentication service.
type AuthService interface {
	Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, error)
}

// authService implements AuthService.
type authService struct {
	userRepo               repository.UserRepository
	jwtMgr                 *jwt.JWTManager
	deviceFingerprintService devicefingerprintservice.DeviceFingerprintService
}

// NewAuthService creates a new AuthService.
func NewAuthService(userRepo repository.UserRepository, jwtMgr *jwt.JWTManager, deviceFPService devicefingerprintservice.DeviceFingerprintService) AuthService {
	return &authService{
		userRepo:               userRepo,
		jwtMgr:                 jwtMgr,
		deviceFingerprintService: deviceFPService,
	}
}

// Login authenticates a user and returns JWT tokens.
func (s *authService) Login(ctx context.Context, req *model.LoginRequest) (*model.LoginResponse, error) {
	if req == nil {
		return nil, errors.New("login request cannot be nil")
	}

	// Identify credential type
	credType := util.IdentifyCredentialType(req.Credential)
	var user *model.User
	var err error

	// Get user based on credential type
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
		return nil, errors.New("invalid credentials")
	}

	// For now, only support password login
	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check device fingerprint if provided
	if req.DeviceFingerprintID != "" {
		// Verify device is trusted
		isTrusted, err := s.deviceFingerprintService.IsTrusted(ctx, user.ID, req.DeviceFingerprintID)
		if err != nil {
			return nil, fmt.Errorf("failed to verify device trust: %w", err)
		}
		
		// If device is not trusted, we would typically require additional verification
		// For this implementation, we'll continue but in a real system you might want to
		// trigger 2FA or other verification methods here
		if !isTrusted {
			// Device is not trusted - log this for security monitoring
			// In a production system, you might want to require additional verification
			// For now, we'll continue with login but this would be where you'd enforce 2FA
		}
	}

	// Generate token pair
	accessToken, refreshToken, err := s.jwtMgr.GenerateTokenPair(user.ID, user.AccountID)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(time.Hour * 24), // 24 hours for access token
		UserID:       user.ID,
		AccountID:    user.AccountID,
	}, nil
}

// RefreshToken generates a new access token using a refresh token.
func (s *authService) RefreshToken(ctx context.Context, refreshToken string) (*model.LoginResponse, error) {
	// Validate refresh token and get new access token
	accessToken, err := s.jwtMgr.RefreshAccessToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// We need to get user info from the refresh token to populate response
	claims, err := s.jwtMgr.ValidateToken(refreshToken)
	if err != nil {
		return nil, err
	}

	return &model.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken, // Return the same refresh token
		ExpiresIn:    int64(time.Hour * 24), // 24 hours for access token
		UserID:       claims.UserID,
		AccountID:    claims.AccountID,
	}, nil
}