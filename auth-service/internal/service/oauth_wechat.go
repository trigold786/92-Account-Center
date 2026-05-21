package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type WeChatOAuthProvider struct {
	appID       string
	appSecret   string
	redirectURI string
}

func NewWeChatOAuthProvider(appID, appSecret, redirectURI string) *WeChatOAuthProvider {
	return &WeChatOAuthProvider{appID: appID, appSecret: appSecret, redirectURI: redirectURI}
}

func (p *WeChatOAuthProvider) Name() string { return "wechat" }

func (p *WeChatOAuthProvider) GetAuthURL(state string) string {
	return fmt.Sprintf("https://open.weixin.qq.com/connect/qrconnect?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_login&state=%s#wechat_redirect",
		p.appID, p.redirectURI, state)
}

func (p *WeChatOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	return "wechat_mock_" + hex.EncodeToString(b), nil
}

func (p *WeChatOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	return &model.SocialUserInfo{
		Provider:    "wechat",
		ProviderUID: "wx_mock_" + accessToken[len(accessToken)-8:],
		Name:        "WeChatUser",
		AvatarURL:   "https://example.com/avatar.png",
	}, nil
}
