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

type APNsConfig struct {
	Mode        string
	Endpoint    string
	BundleID    string
	AccessToken string
}

func (c APNsConfig) ValidateProduction() error {
	if c.Mode != "production" {
		return nil
	}
	fields := map[string]string{
		"APNS_BUNDLE_ID":    c.BundleID,
		"APNS_ACCESS_TOKEN": c.AccessToken,
	}
	for name, value := range fields {
		if insecurePushValue(value) {
			return fmt.Errorf("production apns push config requires real %s", name)
		}
	}
	return nil
}

type apnsProvider struct {
	cfg APNsConfig
}

func NewAPNsProvider(cfg APNsConfig) PushProvider {
	return &apnsProvider{cfg: cfg}
}

func (p *apnsProvider) Name() string { return "apns" }

func (p *apnsProvider) Send(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	if p.cfg.BundleID == "" || p.cfg.AccessToken == "" {
		return nil, fmt.Errorf("apns provider not configured: missing bundle_id or access_token")
	}

	endpoint := p.cfg.Endpoint
	if endpoint == "" {
		endpoint = "https://api.push.apple.com"
	}

	aps := map[string]interface{}{
		"alert": map[string]interface{}{
			"title": req.Title,
			"body":  req.Body,
		},
	}
	if req.Sound != "" {
		aps["sound"] = req.Sound
	}
	if req.Badge != nil {
		aps["badge"] = *req.Badge
	}

	payload := map[string]interface{}{
		"aps": aps,
	}
	for k, v := range req.Data {
		payload[k] = v
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("apns marshal: %w", err)
	}

	url := fmt.Sprintf("%s/3/device/%s", endpoint, req.DeviceToken)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("apns request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("apns-topic", p.cfg.BundleID)
	httpReq.Header.Set("Authorization", "Bearer "+p.cfg.AccessToken)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("apns send: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))

	if resp.StatusCode != http.StatusOK {
		return &PushResponse{
			Success: false,
			Error:   fmt.Sprintf("apns error: status %d: %s", resp.StatusCode, string(respBody)),
		}, nil
	}

	return &PushResponse{
		Success:   true,
		MessageID: resp.Header.Get("apns-id"),
	}, nil
}

func (p *apnsProvider) ValidateToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("apns: device token is empty")
	}
	return nil
}
