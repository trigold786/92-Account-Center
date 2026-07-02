package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

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
	return fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=openid%%20email%%20profile&state=%s&access_type=offline",
		p.clientID, p.redirectURI, state)
}

func (p *GoogleOAuthProvider) ExchangeCode(ctx context.Context, code string) (string, error) {
	data := url.Values{
		"code":          {code},
		"client_id":     {p.clientID},
		"client_secret": {p.clientSecret},
		"redirect_uri":  {p.redirectURI},
		"grant_type":    {"authorization_code"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://oauth2.googleapis.com/token",
		strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("google token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("google token exchange failed (status %d): %s", resp.StatusCode, string(body))
	}

	var result struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("google token response parse failed: %w", err)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("google returned empty access_token")
	}
	return result.AccessToken, nil
}

func (p *GoogleOAuthProvider) GetUserInfo(ctx context.Context, accessToken string) (*model.SocialUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		"https://www.googleapis.com/oauth2/v2/userinfo", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google userinfo request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo failed (status %d): %s", resp.StatusCode, string(body))
	}

	var info struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, fmt.Errorf("google userinfo parse failed: %w", err)
	}
	if info.ID == "" {
		return nil, fmt.Errorf("google returned empty user id")
	}

	return &model.SocialUserInfo{
		Provider:    "google",
		ProviderUID: info.ID,
		Email:       info.Email,
		Name:        info.Name,
		AvatarURL:   info.Picture,
	}, nil
}
