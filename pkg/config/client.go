package config

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

type configItem struct {
	Code         string `json:"code"`
	CurrentValue string `json:"current_value"`
	IsSensitive  bool   `json:"is_sensitive"`
}

type apiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    *configItem `json:"data"`
}

func NewClient(configServiceURL string) *Client {
	return &Client{
		baseURL: configServiceURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (c *Client) GetConfig(code string) (string, error) {
	url := fmt.Sprintf("%s/internal/v1/config/items/%s", c.baseURL, code)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("config client: failed to fetch %s: %w", code, err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("config client: failed to read response: %w", err)
	}

	var apiResp apiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return "", fmt.Errorf("config client: failed to parse response: %w", err)
	}

	if apiResp.Code != 0 {
		return "", fmt.Errorf("config client: API error for %s: %s", code, apiResp.Message)
	}

	if apiResp.Data == nil {
		return "", fmt.Errorf("config client: config %s not found", code)
	}

	return apiResp.Data.CurrentValue, nil
}

func (c *Client) GetConfigInt(code string) (int, error) {
	val, err := c.GetConfig(code)
	if err != nil {
		return 0, err
	}
	var result int
	if _, err := fmt.Sscanf(val, "%d", &result); err != nil {
		return 0, fmt.Errorf("config client: failed to parse %s as int: %w", code, err)
	}
	return result, nil
}

func (c *Client) GetConfigBool(code string) (bool, error) {
	val, err := c.GetConfig(code)
	if err != nil {
		return false, err
	}
	return val == "true", nil
}

func (c *Client) GetConfigDuration(code string) (time.Duration, error) {
	val, err := c.GetConfig(code)
	if err != nil {
		return 0, err
	}
	d, err := parseDuration(val)
	if err != nil {
		return 0, fmt.Errorf("config client: failed to parse %s as duration: %w", code, err)
	}
	return d, nil
}

func parseDuration(s string) (time.Duration, error) {
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration: %q", s)
	}
	if s[len(s)-1] == 'd' {
		var days int
		if _, err := fmt.Sscanf(s, "%dd", &days); err != nil {
			return 0, fmt.Errorf("invalid duration: %q", s)
		}
		return time.Duration(days) * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func (c *Client) GetConfigFloat(code string) (float64, error) {
	val, err := c.GetConfig(code)
	if err != nil {
		return 0, err
	}
	var result float64
	if _, err := fmt.Sscanf(val, "%f", &result); err != nil {
		return 0, fmt.Errorf("config client: failed to parse %s as float: %w", code, err)
	}
	return result, nil
}
