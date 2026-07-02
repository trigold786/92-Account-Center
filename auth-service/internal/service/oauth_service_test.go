package service

import (
	"context"
	"errors"
	"testing"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type mockOAuthProvider struct {
	name     string
	authURL  string
	userInfo *model.SocialUserInfo
}

func (m *mockOAuthProvider) Name() string { return m.name }

func (m *mockOAuthProvider) GetAuthURL(state string) string {
	return m.authURL + "?state=" + state
}

func (m *mockOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	if code == "fail" {
		return "", errors.New("exchange failed")
	}
	return "mock_token_" + code, nil
}

func (m *mockOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	if accessToken == "bad_token" {
		return nil, errors.New("invalid token")
	}
	return m.userInfo, nil
}

func TestOAuthProviderRegistry(t *testing.T) {
	reg := NewOAuthProviderRegistry()
	mock := &mockOAuthProvider{name: "mock_test", authURL: "https://example.com/oauth"}
	reg.Register(mock)

	got, err := reg.Get("mock_test")
	if err != nil {
		t.Fatalf("expected provider, got error: %v", err)
	}
	if got.Name() != "mock_test" {
		t.Fatalf("unexpected name: %s", got.Name())
	}

	_, err = reg.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
}

func TestOAuthFlow(t *testing.T) {
	mock := &mockOAuthProvider{
		name:    "mock",
		authURL: "https://example.com/oauth",
		userInfo: &model.SocialUserInfo{
			Provider:    "mock",
			ProviderUID: "uid_123",
			Email:       "test@example.com",
			AvatarURL:   "https://example.com/avatar.png",
		},
	}
	svc := &OAuthService{
		registry: NewOAuthProviderRegistry(),
		userRepo: &mockUserRepo{},
		roleRepo: &mockRoleRepo{},
	}
	svc.registry.Register(mock)

	url, err := svc.GetAuthURL(context.Background(), "mock", "state_abc")
	if err != nil {
		t.Fatalf("GetAuthURL failed: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}

	_, err = svc.HandleCallback(context.Background(), "mock", "fail")
	if err == nil {
		t.Fatal("expected error for failing code exchange")
	}
}

func TestOAuthProviderNotFound(t *testing.T) {
	svc := &OAuthService{
		registry: NewOAuthProviderRegistry(),
		userRepo: &mockUserRepo{},
	}
	_, err := svc.HandleCallback(context.Background(), "nonexistent", "code")
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
}
