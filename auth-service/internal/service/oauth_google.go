package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type GoogleOAuthProvider struct {
	clientID     string
	clientSecret string
	redirectURI  string
}

func NewGoogleOAuthProvider(clientID, clientSecret, redirectURI string) *GoogleOAuthProvider {
	return &GoogleOAuthProvider{clientID: clientID, clientSecret: clientSecret, redirectURI: redirectURI}
}

func (p *GoogleOAuthProvider) Name() string { return "google" }

func (p *GoogleOAuthProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid%%20email%%20profile&state=%s",
		p.clientID, p.redirectURI, state)
}

func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	return "google_mock_" + hex.EncodeToString(b), nil
}

func (p *GoogleOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	return &model.SocialUserInfo{
		Provider:    "google",
		ProviderUID: "google_mock_" + accessToken[len(accessToken)-8:],
		Name:        "GoogleUser",
		Email:       "user@gmail.com",
		AvatarURL:   "https://example.com/avatar.png",
	}, nil
}
