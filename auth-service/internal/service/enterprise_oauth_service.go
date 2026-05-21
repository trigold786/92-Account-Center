package service

import (
	"context"
	"errors"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type EnterpriseOAuthProvider interface {
	Name() string
	GetAuthURL(state string) string
	ExchangeCode(ctx context.Context, code string) (string, error)
	GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error)
}

type EnterpriseOAuthService struct {
	providers map[string]EnterpriseOAuthProvider
}

func NewEnterpriseOAuthService() *EnterpriseOAuthService {
	return &EnterpriseOAuthService{
		providers: make(map[string]EnterpriseOAuthProvider),
	}
}

func (s *EnterpriseOAuthService) Register(p EnterpriseOAuthProvider) {
	s.providers[p.Name()] = p
}

func (s *EnterpriseOAuthService) GetAuthURL(ctx context.Context, providerName, state string) (string, error) {
	p, ok := s.providers[providerName]
	if !ok {
		return "", errors.New("enterprise oauth provider not found: " + providerName)
	}
	return p.GetAuthURL(state), nil
}

func (s *EnterpriseOAuthService) HandleCallback(ctx context.Context, providerName, code string) (string, *model.SocialUserInfo, error) {
	p, ok := s.providers[providerName]
	if !ok {
		return "", nil, errors.New("enterprise oauth provider not found: " + providerName)
	}
	token, err := p.ExchangeCode(ctx, code)
	if err != nil {
		return "", nil, err
	}
	info, err := p.GetUserInfo(ctx, token)
	if err != nil {
		return "", nil, err
	}
	return token, info, nil
}
