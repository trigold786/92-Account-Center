package provider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type DingTalkProvider struct {
	appKey      string
	appSecret   string
	redirectURI string
}

func NewDingTalkProvider(appKey, appSecret, redirectURI string) *DingTalkProvider {
	return &DingTalkProvider{
		appKey:      appKey,
		appSecret:   appSecret,
		redirectURI: redirectURI,
	}
}

func (p *DingTalkProvider) Name() string { return "dingtalk" }

func (p *DingTalkProvider) GetAuthURL(state string) string {
	return fmt.Sprintf(
		"https://login.dingtalk.com/oauth2/auth?redirect_uri=%s&response_type=code&client_id=%s&scope=openid&state=%s&prompt=consent",
		p.redirectURI, p.appKey, state,
	)
}

func (p *DingTalkProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	return "dingtalk_" + hex.EncodeToString(b), nil
}

func (p *DingTalkProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	return &model.SocialUserInfo{
		Provider:    "dingtalk",
		ProviderUID: "dt_" + accessToken[len(accessToken)-8:],
		Name:        "DingTalkUser",
		AvatarURL:   "https://example.com/dt_avatar.png",
	}, nil
}
