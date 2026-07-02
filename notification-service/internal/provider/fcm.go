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

type FCMConfig struct {
	Mode        string
	Endpoint    string
	ProjectID   string
	AccessToken string
}

func (c FCMConfig) ValidateProduction() error {
	if c.Mode != "production" {
		return nil
	}
	fields := map[string]string{
		"FCM_PROJECT_ID":   c.ProjectID,
		"FCM_ACCESS_TOKEN": c.AccessToken,
	}
	for name, value := range fields {
		if insecurePushValue(value) {
			return fmt.Errorf("production fcm push config requires real %s", name)
		}
	}
	return nil
}

type fcmProvider struct {
	cfg FCMConfig
}

func NewFCMProvider(cfg FCMConfig) PushProvider {
	return &fcmProvider{cfg: cfg}
}

func (p *fcmProvider) Name() string { return "fcm" }

func (p *fcmProvider) Send(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	if p.cfg.ProjectID == "" || p.cfg.AccessToken == "" {
		return nil, fmt.Errorf("fcm provider not configured: missing project_id or access_token")
	}

	endpoint := p.cfg.Endpoint
	if endpoint == "" {
		endpoint = "https://fcm.googleapis.com"
	}

	payload := map[string]interface{}{
		"message": map[string]interface{}{
			"token": req.DeviceToken,
			"notification": map[string]interface{}{
				"title": req.Title,
				"body":  req.Body,
			},
			"data": req.Data,
		},
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("fcm marshal: %w", err)
	}

	url := fmt.Sprintf("%s/v1/projects/%s/messages:send", endpoint, p.cfg.ProjectID)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("fcm request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.cfg.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("fcm send: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusOK {
		return &PushResponse{
			Success: false,
			Error:   fmt.Sprintf("fcm error: status %d: %s", resp.StatusCode, string(respBody)),
		}, nil
	}

	var result struct {
		Name string `json:"name"`
	}
	json.Unmarshal(respBody, &result)

	return &PushResponse{
		Success:   true,
		MessageID: result.Name,
	}, nil
}

func (p *fcmProvider) ValidateToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("fcm: device token is empty")
	}
	return nil
}
