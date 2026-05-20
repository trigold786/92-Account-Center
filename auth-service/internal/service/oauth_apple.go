package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type AppleOAuthProvider struct {
	clientID    string
	teamID      string
	keyID       string
	redirectURI string
}

func NewAppleOAuthProvider(clientID, teamID, keyID, redirectURI string) *AppleOAuthProvider {
	return &AppleOAuthProvider{clientID: clientID, teamID: teamID, keyID: keyID, redirectURI: redirectURI}
}

func (p *AppleOAuthProvider) Name() string { return "apple" }

func (p *AppleOAuthProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://appleid.apple.com/auth/authorize?client_id=%s&redirect_uri=%s&response_type=code%%20id_token&scope=name%%20email&state=%s&response_mode=form_post",
		p.clientID, p.redirectURI, state)
}

func (p *AppleOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	return "apple_mock_" + hex.EncodeToString(b), nil
}

func (p *AppleOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	return &model.SocialUserInfo{
		Provider:    "apple",
		ProviderUID: "apple_mock_" + accessToken[len(accessToken)-8:],
		Email:       "user@privaterelay.appleid.com",
	}, nil
}
