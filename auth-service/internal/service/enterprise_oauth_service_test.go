package service

import (
	"context"
	"testing"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
)

type mockEnterpriseProvider struct {
	name     string
	authURL  string
	userInfo *model.SocialUserInfo
}

func (m *mockEnterpriseProvider) Name() string { return m.name }
func (m *mockEnterpriseProvider) GetAuthURL(state string) string {
	return m.authURL + "&state=" + state
}
func (m *mockEnterpriseProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	return "ent_token_" + code, nil
}
func (m *mockEnterpriseProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	return m.userInfo, nil
}

func TestWorkWeChatAuthURL(t *testing.T) {
	svc := NewEnterpriseOAuthService()
	mock := &mockEnterpriseProvider{
		name:    "work_wechat",
		authURL: "https://open.work.weixin.qq.com/wwopen/sso/qrConnect?appid=test",
		userInfo: &model.SocialUserInfo{
			Provider:    "work_wechat",
			ProviderUID: "ww_123",
			Name:        "TestUser",
		},
	}
	svc.Register(mock)

	url, err := svc.GetAuthURL(context.Background(), "work_wechat", "test_state")
	if err != nil {
		t.Fatalf("GetAuthURL failed: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
}

func TestDingTalkAuthURL(t *testing.T) {
	svc := NewEnterpriseOAuthService()
	mock := &mockEnterpriseProvider{
		name:    "dingtalk",
		authURL: "https://login.dingtalk.com/oauth2/auth?client_id=test",
		userInfo: &model.SocialUserInfo{
			Provider:    "dingtalk",
			ProviderUID: "dt_456",
			Name:        "DingUser",
		},
	}
	svc.Register(mock)

	url, err := svc.GetAuthURL(context.Background(), "dingtalk", "dt_state")
	if err != nil {
		t.Fatalf("GetAuthURL failed: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
}

func TestHandleCallback(t *testing.T) {
	svc := NewEnterpriseOAuthService()
	mock := &mockEnterpriseProvider{
		name:    "work_wechat",
		authURL: "https://example.com",
		userInfo: &model.SocialUserInfo{
			Provider:    "work_wechat",
			ProviderUID: "ww_cb_123",
			Name:        "CallbackUser",
		},
	}
	svc.Register(mock)

	token, info, err := svc.HandleCallback(context.Background(), "work_wechat", "valid_code")
	if err != nil {
		t.Fatalf("HandleCallback failed: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
	if info.Provider != "work_wechat" {
		t.Fatalf("expected provider 'work_wechat', got %s", info.Provider)
	}

	_, _, err = svc.HandleCallback(context.Background(), "nonexistent", "code")
	if err == nil {
		t.Fatal("expected error for nonexistent provider")
	}
}
