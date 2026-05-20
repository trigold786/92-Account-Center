//go:build integration

package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

func getBaseURL() string {
	return getEnv("GATEWAY_URL", "http://localhost:30300")
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getServiceURL(envKey, fallback string) string {
	return getEnv(envKey, fallback)
}

func apiRequest(t *testing.T, method, path string, body interface{}, token string) (*http.Response, map[string]interface{}) {
	t.Helper()

	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	url := getBaseURL() + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("failed to create request %s %s: %v", method, path, err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, path, err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &result); err != nil {
			t.Fatalf("failed to parse response JSON from %s %s: %v (body: %s)", method, path, err, string(respBody))
		}
	}

	return resp, result
}

func rawAPIRequest(method, path string, body interface{}, token string) (*http.Response, []byte, error) {
	var reqBody io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, nil, err
		}
		reqBody = bytes.NewReader(jsonBytes)
	}

	url := getBaseURL() + path
	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, respBody, nil
}

func registerUser(t *testing.T, phone, accountID, password string) (map[string]interface{}, string) {
	t.Helper()
	resp, body := apiRequest(t, http.MethodPost, "/api/v1/account/register", map[string]interface{}{
		"phone_number":  phone,
		"account_id":    accountID,
		"password":      password,
		"agree_to_terms": true,
	}, "")
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("register failed with status %d: %v", resp.StatusCode, body)
	}
	userID := ""
	if id, ok := body["id"]; ok {
		userID = fmt.Sprintf("%v", id)
	}
	return body, userID
}

func loginUser(t *testing.T, credential, password string) (map[string]interface{}, string) {
	t.Helper()
	resp, body := apiRequest(t, http.MethodPost, "/api/v1/auth/login", map[string]interface{}{
		"credential": credential,
		"password":   password,
	}, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login failed with status %d: %v", resp.StatusCode, body)
	}
	token, _ := body["access_token"].(string)
	return body, token
}

func refreshToken(t *testing.T, refreshTokenStr string) (map[string]interface{}, string) {
	t.Helper()
	resp, body := apiRequest(t, http.MethodPost, "/api/v1/auth/refresh", map[string]interface{}{
		"refresh_token": refreshTokenStr,
	}, "")
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("token refresh failed with status %d: %v", resp.StatusCode, body)
	}
	token, _ := body["access_token"].(string)
	return body, token
}

func getUserTier(t *testing.T, token, userID string) map[string]interface{} {
	t.Helper()
	resp, body := apiRequest(t, http.MethodGet, "/api/v1/account/"+userID+"/tier", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get tier failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func purchaseSubscription(t *testing.T, token, userID string, tier int) map[string]interface{} {
	t.Helper()
	resp, body := apiRequest(t, http.MethodPost, "/api/v1/subscriptions/purchase", map[string]interface{}{
		"user_id":    userID,
		"tier_level": tier,
		"price":      29.90,
	}, token)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		t.Fatalf("purchase subscription failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func getUserSubscriptions(t *testing.T, token, userID string) map[string]interface{} {
	t.Helper()
	resp, body := apiRequest(t, http.MethodGet, "/api/v1/subscriptions/"+userID, nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get subscriptions failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func getCredits(t *testing.T, token, userID string) map[string]interface{} {
	t.Helper()
	resp, body := apiRequest(t, http.MethodGet, "/api/v1/credits/"+userID+"/account", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get credits failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func getCreditTransactions(t *testing.T, token, userID string) map[string]interface{} {
	t.Helper()
	resp, body := apiRequest(t, http.MethodGet, "/api/v1/credits/"+userID+"/transactions", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get credit transactions failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func getEntitlements(t *testing.T, token, userID string) map[string]interface{} {
	t.Helper()
	resp, body := apiRequest(t, http.MethodGet, "/api/v1/entitlements/"+userID, nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get entitlements failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func requestDeletion(t *testing.T, token string) map[string]interface{} {
	t.Helper()
	_, body := apiRequest(t, http.MethodPost, "/api/v1/account/deletion/request", map[string]interface{}{
		"verification_code": "123456",
		"verification_type": "sms_code",
	}, token)
	return body
}

func getDeletionStatus(t *testing.T, token string) map[string]interface{} {
	t.Helper()
	_, body := apiRequest(t, http.MethodGet, "/api/v1/account/deletion/status", nil, token)
	return body
}

func cancelDeletion(t *testing.T, token string) map[string]interface{} {
	t.Helper()
	_, body := apiRequest(t, http.MethodPost, "/api/v1/account/deletion/cancel", nil, token)
	return body
}

func bindReferral(t *testing.T, token, referralCode string) map[string]interface{} {
	t.Helper()
	_, body := apiRequest(t, http.MethodPost, "/api/v1/referral/bind", map[string]interface{}{
		"referral_code": referralCode,
	}, token)
	return body
}

func generateReferralLink(t *testing.T, token string) map[string]interface{} {
	t.Helper()
	_, body := apiRequest(t, http.MethodPost, "/api/v1/referral/generate-link", nil, token)
	return body
}

func getReferralSummary(t *testing.T, token, userID string) map[string]interface{} {
	t.Helper()
	resp, body := apiRequest(t, http.MethodGet, "/api/v1/referral/"+userID+"/summary", nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get referral summary failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func createOrder(t *testing.T, token string, userID int64, productType, productName string, amount float64) map[string]interface{} {
	t.Helper()
	resp, body := apiRequest(t, http.MethodPost, "/api/v1/orders", map[string]interface{}{
		"user_id":      userID,
		"product_type": productType,
		"product_name": productName,
		"amount":       amount,
		"currency":     "CNY",
	}, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("create order failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func getOrder(t *testing.T, token, orderID string) map[string]interface{} {
	t.Helper()
	resp, body := apiRequest(t, http.MethodGet, "/api/v1/orders/"+orderID, nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("get order failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func updateOrderStatus(t *testing.T, token, orderID, status string) map[string]interface{} {
	t.Helper()
	resp, body := apiRequest(t, http.MethodPut, "/api/v1/orders/"+orderID+"/status", map[string]interface{}{
		"status":                 status,
		"payment_method":         "alipay",
		"payment_transaction_id": "txn-test-001",
	}, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("update order status failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func listOrders(t *testing.T, token string, userID int64) map[string]interface{} {
	t.Helper()
	url := fmt.Sprintf("/api/v1/orders?user_id=%d&page=1&page_size=10", userID)
	resp, body := apiRequest(t, http.MethodGet, url, nil, token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("list orders failed with status %d: %v", resp.StatusCode, body)
	}
	return body
}

func waitForCondition(t *testing.T, timeout time.Duration, condition func() bool, msg string) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	interval := 500 * time.Millisecond
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(interval)
	}
	t.Fatalf("timed out waiting for condition: %s", msg)
}

func waitForServices(t *testing.T, timeout time.Duration) {
	t.Helper()
	services := []struct {
		name string
		url  string
	}{
		{"api-gateway", getBaseURL() + "/health"},
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		allReady := true
		for _, svc := range services {
			resp, err := http.Get(svc.url)
			if err != nil || resp.StatusCode != 200 {
				allReady = false
				break
			}
			if resp.Body != nil {
				_ = resp.Body.Close()
			}
		}
		if allReady {
			return
		}
		time.Sleep(1 * time.Second)
	}
	t.Fatal("timed out waiting for services to be ready")
}

func assertStatus(t *testing.T, body map[string]interface{}, expectedKey string, expectedValue interface{}) {
	t.Helper()
	val, ok := body[expectedKey]
	if !ok {
		t.Fatalf("response missing key %q: %v", expectedKey, body)
	}
	if val != expectedValue {
		t.Fatalf("expected %s=%v, got %v (body: %v)", expectedKey, expectedValue, val, body)
	}
}

func assertKeyExists(t *testing.T, body map[string]interface{}, key string) {
	t.Helper()
	if _, ok := body[key]; !ok {
		t.Fatalf("response missing key %q: %v", key, body)
	}
}
