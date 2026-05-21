package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type AlipayOAuthProvider struct {
	appID       string
	privateKey  string
	redirectURI string
}

func NewAlipayOAuthProvider(appID, privateKey, redirectURI string) *AlipayOAuthProvider {
	return &AlipayOAuthProvider{appID: appID, privateKey: privateKey, redirectURI: redirectURI}
}

func (p *AlipayOAuthProvider) Name() string { return "alipay" }

func (p *AlipayOAuthProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://openauth.alipay.com/oauth2/publicAppAuthorize.htm?app_id=%s&redirect_uri=%s&scope=auth_user&state=%s",
		p.appID, p.redirectURI, state)
}

func (p *AlipayOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	return "alipay_mock_" + hex.EncodeToString(b), nil
}

func (p *AlipayOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	return &model.SocialUserInfo{
		Provider:    "alipay",
		ProviderUID: "alipay_mock_" + accessToken[len(accessToken)-8:],
		Name:        "AlipayUser",
		AvatarURL:   "https://example.com/alipay_avatar.png",
	}, nil
}
