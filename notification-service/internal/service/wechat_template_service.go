package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

// WeChatTemplateService sends WeChat subscription messages using template IDs.
type WeChatTemplateService struct {
	appID          string
	appSecret      string
	templateIDs    map[string]string
	httpClient     *http.Client
	logger         *slog.Logger

	mu            sync.RWMutex
	accessToken   string
	tokenExpireAt time.Time
}

func NewWeChatTemplateService(appID, appSecret string, templateIDs map[string]string) *WeChatTemplateService {
	return &WeChatTemplateService{
		appID:       appID,
		appSecret:   appSecret,
		templateIDs: templateIDs,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		logger:      slog.Default(),
	}
}

var validTemplates = map[string]bool{
	"subscription_expiring": true,
	"payment_success":       true,
	"referral_bonus":        true,
	"tier_upgrade":          true,
}

func (s *WeChatTemplateService) ValidateTemplateID(templateType string) error {
	if !validTemplates[templateType] {
		return errors.New("invalid template type: " + templateType)
	}
	if _, ok := s.templateIDs[templateType]; !ok {
		return fmt.Errorf("no template ID configured for type: %s", templateType)
	}
	return nil
}

func (s *WeChatTemplateService) getAccessToken(ctx context.Context) (string, error) {
	s.mu.RLock()
	if s.accessToken != "" && time.Now().Before(s.tokenExpireAt) {
		defer s.mu.RUnlock()
		return s.accessToken, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.accessToken != "" && time.Now().Before(s.tokenExpireAt) {
		return s.accessToken, nil
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", s.appID, s.appSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch access token: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("parse access token response: %w", err)
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("wechat access token error %d: %s", result.ErrCode, result.ErrMsg)
	}
	if result.AccessToken == "" {
		return "", fmt.Errorf("empty access token from wechat")
	}

	s.accessToken = result.AccessToken
	s.tokenExpireAt = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)
	return s.accessToken, nil
}

func (s *WeChatTemplateService) SendSubscriptionMessage(ctx context.Context, openID, templateType string, data map[string]string) error {
	if err := s.ValidateTemplateID(templateType); err != nil {
		return err
	}

	accessToken, err := s.getAccessToken(ctx)
	if err != nil {
		return fmt.Errorf("get access token: %w", err)
	}

	templateID := s.templateIDs[templateType]

	type templateDataItem struct {
		Value string `json:"value"`
	}
	fields := make(map[string]templateDataItem)
	for k, v := range data {
		fields[k] = templateDataItem{Value: v}
	}

	payload := map[string]interface{}{
		"touser":      openID,
		"template_id": templateID,
		"data":        fields,
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", accessToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payloadBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send subscribe message: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("parse send response: %w", err)
	}

	switch result.ErrCode {
	case 0:
		s.logger.Info("wechat subscribe message sent", "open_id", openID, "type", templateType)
		return nil
	case 43004:
		return fmt.Errorf("user %s has unsubscribed", openID)
	case 43101:
		return fmt.Errorf("user %s is not subscribed", openID)
	case 40001:
		s.accessToken = ""
		s.tokenExpireAt = time.Time{}
		return fmt.Errorf("access token expired, retrying")
	case 41028:
		return fmt.Errorf("template ID %s not found in user subscriptions", templateID)
	case 41029:
		return fmt.Errorf("template request rejected by user")
	default:
		return fmt.Errorf("wechat error %d: %s", result.ErrCode, result.ErrMsg)
	}
}

func (s *WeChatTemplateService) SendWithRetry(ctx context.Context, openID, templateType string, data map[string]string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = s.SendSubscriptionMessage(ctx, openID, templateType, data)
		if err == nil {
			return nil
		}
		if errors.Is(err, context.Canceled) {
			return err
		}
		if i < 2 {
			s.logger.Warn("retrying template message", "attempt", i+1, "error", err)
			time.Sleep(time.Duration(1<<uint(i)) * time.Second)
		}
	}
	return fmt.Errorf("failed after 3 retries: %w", err)
}
