package service

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

var (
	ErrProviderNotFound = errors.New("oauth provider not found")
	ErrCodeExchange     = errors.New("code exchange failed")
)

type OAuthProvider interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (string, error)
	GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error)
}

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

type OAuthService struct {
	registry *OAuthProviderRegistry
	userRepo UserRepository
}

func NewOAuthService(registry *OAuthProviderRegistry, userRepo UserRepository) *OAuthService {
	return &OAuthService{registry: registry, userRepo: userRepo}
}

func (s *OAuthService) GetAuthURL(ctx context.Context, providerName, state string) (string, error) {
	p, err := s.registry.Get(providerName)
	if err != nil {
		return "", err
	}
	return p.GetAuthURL(state), nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, providerName, code string) (accessToken string, userInfo *model.SocialUserInfo, err error) {
	p, err := s.registry.Get(providerName)
	if err != nil {
		return "", nil, err
	}
	token, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return "", nil, fmt.Errorf("%w: %v", ErrCodeExchange, err)
	}
	info, err := p.GetUserInfo(ctx, token)
	if err != nil {
		return "", nil, err
	}
	return token, info, nil
}
