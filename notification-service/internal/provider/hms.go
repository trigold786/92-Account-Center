package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type HMSConfig struct {
	Mode        string
	Endpoint    string
	AppID       string
	AccessToken string
}

func (c HMSConfig) ValidateProduction() error {
	if c.Mode != "production" {
		return nil
	}
	fields := map[string]string{
		"HMS_APP_ID":       c.AppID,
		"HMS_ACCESS_TOKEN": c.AccessToken,
	}
	for name, value := range fields {
		if insecurePushValue(value) {
			return fmt.Errorf("production hms push config requires real %s", name)
		}
	}
	return nil
}

type hmsProvider struct {
	cfg HMSConfig
}

func NewHMSProvider(cfg HMSConfig) PushProvider {
	return &hmsProvider{cfg: cfg}
}

func (p *hmsProvider) Name() string { return "hms" }

func (p *hmsProvider) Send(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	if p.cfg.AppID == "" || p.cfg.AccessToken == "" {
		return nil, fmt.Errorf("hms provider not configured: missing app_id or access_token")
	}

	endpoint := p.cfg.Endpoint
	if endpoint == "" {
		endpoint = "https://push-api.cloud.hicloud.com"
	}

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"android": map[string]interface{}{
				"notification": map[string]interface{}{
					"title": req.Title,
					"body":  req.Body,
				},
			},
			"token": []string{req.DeviceToken},
			"data":  req.Data,
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("hms marshal: %w", err)
	}

	url := fmt.Sprintf("%s/v2/%s/message:send", endpoint, p.cfg.AppID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("hms request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.cfg.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("hms send: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusOK {
		return &PushResponse{
			Success: false,
			Error:   fmt.Sprintf("hms error: status %d: %s", resp.StatusCode, string(respBody)),
		}, nil
	}

	var result struct {
		Code      string `json:"code"`
		Msg       string `json:"msg"`
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return &PushResponse{
			Success: false,
			Error:   fmt.Sprintf("hms error: invalid response: %s", string(respBody)),
		}, nil
	}

	if result.Code != "80000000" {
		return &PushResponse{
			Success:   false,
			MessageID: result.RequestID,
			Error:     fmt.Sprintf("hms error: code %s: %s", result.Code, result.Msg),
		}, nil
	}

	return &PushResponse{
		Success:   true,
		MessageID: result.RequestID,
	}, nil
}

func (p *hmsProvider) ValidateToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("hms: device token is empty")
	}
	return nil
}
