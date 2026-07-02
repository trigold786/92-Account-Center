package sms

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	circuitbreaker "github.com/trigold786/92-Account-Center/pkg/circuitbreaker"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: circuitbreaker.WrapHTTPClient(&http.Client{Timeout: 10 * time.Second}, "notification-service-sms"),
	}
}

func (c *Client) SendCode(phoneNumber string) error {
	body, _ := json.Marshal(map[string]string{"phone_number": phoneNumber})
	resp, err := c.httpClient.Post(c.baseURL+"/api/v1/sms/send", "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("sms send request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return fmt.Errorf("sms send failed: %s", string(respBody))
	}
	return nil
}

func (c *Client) VerifyCode(phoneNumber, code string) (bool, error) {
	body, _ := json.Marshal(map[string]string{"phone_number": phoneNumber, "code": code})
	resp, err := c.httpClient.Post(c.baseURL+"/api/v1/sms/verify", "application/json", bytes.NewReader(body))
	if err != nil {
		return false, fmt.Errorf("sms verify request failed: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Data struct {
			Valid bool `json:"valid"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return false, fmt.Errorf("sms verify decode failed: %w", err)
	}
	return result.Data.Valid, nil
}
