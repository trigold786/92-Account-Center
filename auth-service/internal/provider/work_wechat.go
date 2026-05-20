package provider

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type WorkWeChatProvider struct {
	corpID       string
	agentID      string
	corpSecret   string
	redirectURI  string
}

func NewWorkWeChatProvider(corpID, agentID, corpSecret, redirectURI string) *WorkWeChatProvider {
	return &WorkWeChatProvider{
		corpID:      corpID,
		agentID:     agentID,
		corpSecret:  corpSecret,
		redirectURI: redirectURI,
	}
}

func (p *WorkWeChatProvider) Name() string { return "work_wechat" }

func (p *WorkWeChatProvider) GetAuthURL(state string) string {
	return fmt.Sprintf(
		"https://open.work.weixin.qq.com/wwopen/sso/qrConnect?appid=%s&agentid=%s&redirect_uri=%s&state=%s",
		p.corpID, p.agentID, p.redirectURI, state,
	)
}

func (p *WorkWeChatProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	b := make([]byte, 16)
	rand.Read(b)
	return "work_wechat_" + hex.EncodeToString(b), nil
}

func (p *WorkWeChatProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	return &model.SocialUserInfo{
		Provider:    "work_wechat",
		ProviderUID: "ww_" + accessToken[len(accessToken)-8:],
		Name:        "WorkWeChatUser",
		AvatarURL:   "https://example.com/ww_avatar.png",
	}, nil
}
