package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	circuitbreaker "github.com/trigold786/92-Account-Center/pkg/circuitbreaker"
)

type ReferralClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewReferralClient(baseURL string) *ReferralClient {
	return &ReferralClient{
		baseURL:    baseURL,
		httpClient: circuitbreaker.WrapHTTPClient(&http.Client{Timeout: 5 * time.Second}, "credit-service-referral"),
	}
}

func (c *ReferralClient) BindReferral(ctx context.Context, referralCode, refereeID string) error {
	body, _ := json.Marshal(map[string]string{
		"referrer_code": referralCode,
		"referee_id":    refereeID,
	})
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/referral/bind", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("referral bind request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("referral bind failed with status %d", resp.StatusCode)
	}
	return nil
}
