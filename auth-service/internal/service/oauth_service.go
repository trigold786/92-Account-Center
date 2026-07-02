package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
	"github.com/trigold786/92-Account-Center/auth-service/pkg/jwt"
)

var (
	ErrProviderNotFound = errors.New("oauth provider not found")
	ErrCodeExchange     = errors.New("code exchange failed")
)

// OAuthProvider defines the contract for third-party OAuth integrations.
type OAuthProvider interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (string, error)
	GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error)
}

// OAuthProviderRegistry manages registration and lookup of OAuth providers.
type OAuthProviderRegistry struct {
	mu        sync.RWMutex
	providers map[string]OAuthProvider
}

func NewOAuthProviderRegistry() *OAuthProviderRegistry {
	return &OAuthProviderRegistry{
		providers: make(map[string]OAuthProvider),
	}
}

func (r *OAuthProviderRegistry) Register(p OAuthProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.providers[p.Name()] = p
}

func (r *OAuthProviderRegistry) Get(name string) (OAuthProvider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrProviderNotFound, name)
	}
	return p, nil
}

func (r *OAuthProviderRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var names []string
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// OAuthCallbackResult contains the outcome of a successful OAuth callback.
type OAuthCallbackResult struct {
	AccessToken  string  `json:"access_token"`
	RefreshToken string  `json:"refresh_token"`
	UserID       int64   `json:"user_id"`
	AccountID    string  `json:"account_id"`
	Roles        []string `json:"roles"`
	IsNewUser    bool    `json:"is_new_user"`
}

// OAuthService handles OAuth authentication flows and user account creation.
type OAuthService struct {
	registry *OAuthProviderRegistry
	userRepo UserRepository
	roleRepo RoleRepository
	jwtMgr   *jwt.JWTManager
	logger   *slog.Logger
}

// NewOAuthService creates an OAuthService with the given dependencies.
func NewOAuthService(registry *OAuthProviderRegistry, userRepo UserRepository, roleRepo RoleRepository, jwtMgr *jwt.JWTManager) *OAuthService {
	return &OAuthService{
		registry: registry,
		userRepo: userRepo,
		roleRepo: roleRepo,
		jwtMgr:   jwtMgr,
		logger:   slog.Default(),
	}
}

func (s *OAuthService) GetAuthURL(ctx context.Context, providerName, state string) (string, error) {
	p, err := s.registry.Get(providerName)
	if err != nil {
		return "", err
	}
	return p.GetAuthURL(state), nil
}

// HandleCallback exchanges an OAuth code for tokens, creates or links a user, and returns JWTs.
func (s *OAuthService) HandleCallback(ctx context.Context, providerName, code string) (*OAuthCallbackResult, error) {
	p, err := s.registry.Get(providerName)
	if err != nil {
		return nil, err
	}

	providerToken, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCodeExchange, err)
	}

	info, err := p.GetUserInfo(ctx, providerToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.FindBySocialAccount(ctx, providerName, info.ProviderUID)
	if err != nil {
		return nil, fmt.Errorf("find social account: %w", err)
	}

	isNewUser := false
	if user == nil {
		isNewUser = true
		user, err = s.userRepo.CreateFromSocial(ctx, info)
		if err != nil {
			return nil, fmt.Errorf("create user from social: %w", err)
		}
		if err := s.userRepo.LinkSocialAccount(ctx, user.ID, providerName, info.ProviderUID, info.Email, info.AvatarURL, providerToken, ""); err != nil {
			s.logger.Error("failed to link social account", "user_id", user.ID, "provider", providerName, "error", err)
		}
	} else {
		if err := s.userRepo.LinkSocialAccount(ctx, user.ID, providerName, info.ProviderUID, info.Email, info.AvatarURL, providerToken, ""); err != nil {
			s.logger.Error("failed to update social token", "user_id", user.ID, "provider", providerName, "error", err)
		}
	}

	roles, err := s.roleRepo.GetUserRoles(ctx, user.AccountID)
	if err != nil {
		s.logger.Error("failed to get user roles", "account_id", user.AccountID, "error", err)
		roles = []string{}
	}

	tokenResp, err := s.jwtMgr.GenerateTokenPairWithDevice(user.ID, user.AccountID, "", roles)
	if err != nil {
		return nil, fmt.Errorf("generate tokens: %w", err)
	}

	return &OAuthCallbackResult{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		UserID:       user.ID,
		AccountID:    user.AccountID,
		Roles:        roles,
		IsNewUser:    isNewUser,
	}, nil
}
