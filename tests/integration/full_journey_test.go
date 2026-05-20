//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestHealthEndpoints(t *testing.T) {
	services := []struct {
		name string
		url  string
	}{
		{"api-gateway", getServiceURL("GATEWAY_URL", "http://localhost:30300") + "/health"},
		{"account-service", getServiceURL("ACCOUNT_SERVICE_URL", "http://localhost:30301") + "/health"},
		{"auth-service", getServiceURL("AUTH_SERVICE_URL", "http://localhost:30302") + "/health"},
		{"notification-service", getServiceURL("NOTIFICATION_SERVICE_URL", "http://localhost:30311") + "/health"},
		{"credit-service", getServiceURL("CREDIT_SERVICE_URL", "http://localhost:30312") + "/health"},
		{"compliance-service", getServiceURL("COMPLIANCE_SERVICE_URL", "http://localhost:30313") + "/health"},
		{"config-service", getServiceURL("CONFIG_SERVICE_URL", "http://localhost:30315") + "/health"},
	}

	for _, svc := range services {
		t.Run(svc.name, func(t *testing.T) {
			client := &http.Client{Timeout: 10 * time.Second}
			resp, err := client.Get(svc.url)
			if err != nil {
				t.Errorf("%s health check failed: %v", svc.name, err)
				return
			}
			defer func() { _ = resp.Body.Close() }()
			if resp.StatusCode != 200 {
				t.Errorf("%s returned status %d, expected 200", svc.name, resp.StatusCode)
			}
		})
	}
}

func TestAuthenticationFlow(t *testing.T) {
	t.Run("login_with_invalid_credentials", func(t *testing.T) {
		resp, body := apiRequest(t, http.MethodPost, "/api/v1/auth/login", map[string]interface{}{
			"credential": "invalid_user_999",
			"password":   "wrongpassword",
		}, "")
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d: %v", resp.StatusCode, body)
		}
	})

	t.Run("access_protected_without_token", func(t *testing.T) {
		resp, _, err := rawAPIRequest(http.MethodGet, "/api/v1/account/test-user/tier", nil, "")
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 without token, got %d", resp.StatusCode)
		}
	})

	phone := fmt.Sprintf("138%08d", time.Now().UnixNano()%100000000)
	accountID := fmt.Sprintf("authtest%d", time.Now().UnixNano())
	password := "TestAuth123!"

	regBody, userID := registerUser(t, phone, accountID, password)
	assertKeyExists(t, regBody, "id")
	if userID == "" {
		t.Fatal("registration did not return user ID")
	}

	t.Run("login_with_valid_credentials", func(t *testing.T) {
		loginBody, token := loginUser(t, phone, password)
		if token == "" {
			t.Fatal("login succeeded but no access_token returned")
		}
		assertKeyExists(t, loginBody, "refresh_token")
		assertKeyExists(t, loginBody, "user_id")
		assertKeyExists(t, loginBody, "expires_in")
	})

	t.Run("access_protected_with_valid_token", func(t *testing.T) {
		_, token := loginUser(t, phone, password)
		resp, body := apiRequest(t, http.MethodGet, "/api/v1/account/"+userID+"/tier", nil, token)
		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected 200 with valid token, got %d: %v", resp.StatusCode, body)
		}
	})

	t.Run("token_refresh", func(t *testing.T) {
		loginBody, _ := loginUser(t, phone, password)
		refreshTok, ok := loginBody["refresh_token"].(string)
		if !ok || refreshTok == "" {
			t.Fatal("no refresh_token in login response")
		}
		refreshBody, newToken := refreshToken(t, refreshTok)
		if newToken == "" {
			t.Fatal("token refresh succeeded but no new access_token")
		}
		assertKeyExists(t, refreshBody, "access_token")
		assertKeyExists(t, refreshBody, "refresh_token")
	})
}

func TestFullJourney(t *testing.T) {
	phone := fmt.Sprintf("138%08d", time.Now().UnixNano()%100000000)
	accountID := fmt.Sprintf("journey%d", time.Now().UnixNano())
	password := "JourneyTest123!"

	t.Logf("registering user: phone=%s accountID=%s", phone, accountID)

	t.Run("step1_register", func(t *testing.T) {
		regBody, uid := registerUser(t, phone, accountID, password)
		assertKeyExists(t, regBody, "id")
		assertKeyExists(t, regBody, "account_id")
		if uid == "" {
			t.Fatal("no user ID returned from registration")
		}
		t.Logf("registered user ID: %s", uid)
	})

	loginBody, token := loginUser(t, phone, password)
	if token == "" {
		t.Fatal("login failed: no token")
	}
	userIDFloat, ok := loginBody["user_id"].(float64)
	if !ok {
		t.Fatal("login response missing user_id")
	}
	userID := fmt.Sprintf("%d", int64(userIDFloat))
	t.Logf("logged in user ID: %s", userID)

	t.Run("step2_login_and_profile", func(t *testing.T) {
		tierBody := getUserTier(t, token, userID)
		assertKeyExists(t, tierBody, "tier_level")
		t.Logf("initial tier: %v", tierBody["tier_level"])
	})

	t.Run("step3_purchase_subscription", func(t *testing.T) {
		subResp := purchaseSubscription(t, token, userID, 2)
		t.Logf("subscription purchase response: %v", subResp)
	})

	t.Run("step4_verify_subscription_active", func(t *testing.T) {
		subsResp := getUserSubscriptions(t, token, userID)
		t.Logf("subscriptions: %v", subsResp)
	})

	t.Run("step5_check_entitlements", func(t *testing.T) {
		entResp := getEntitlements(t, token, userID)
		t.Logf("entitlements: %v", entResp)
	})

	t.Run("step6_check_credits", func(t *testing.T) {
		creditResp := getCredits(t, token, userID)
		t.Logf("credits: %v", creditResp)
	})

	t.Run("step7_referral_link", func(t *testing.T) {
		linkResp := generateReferralLink(t, token)
		t.Logf("referral link: %v", linkResp)
	})

	t.Run("step8_referral_summary", func(t *testing.T) {
		summaryResp := getReferralSummary(t, token, userID)
		t.Logf("referral summary: %v", summaryResp)
	})

	t.Run("step9_request_deletion", func(t *testing.T) {
		delResp := requestDeletion(t, token)
		t.Logf("deletion response: %v", delResp)
	})

	t.Run("step10_deletion_status", func(t *testing.T) {
		statusResp := getDeletionStatus(t, token)
		t.Logf("deletion status: %v", statusResp)
	})

	t.Run("step11_cancel_deletion", func(t *testing.T) {
		cancelResp := cancelDeletion(t, token)
		t.Logf("cancel deletion response: %v", cancelResp)
	})
}

func TestReferralFlow(t *testing.T) {
	referrerPhone := fmt.Sprintf("139%08d", time.Now().UnixNano()%100000000)
	referrerAccountID := fmt.Sprintf("referrer%d", time.Now().UnixNano())
	password := "ReferralTest123!"

	_, referrerID := registerUser(t, referrerPhone, referrerAccountID, password)
	_, referrerToken := loginUser(t, referrerPhone, password)

	t.Run("generate_referral_link", func(t *testing.T) {
		linkResp := generateReferralLink(t, referrerToken)
		t.Logf("generated referral link: %v", linkResp)
	})

	refereePhone := fmt.Sprintf("137%08d", time.Now().UnixNano()%100000000)
	refereeAccountID := fmt.Sprintf("referee%d", time.Now().UnixNano())
	_, refereeID := registerUser(t, refereePhone, refereeAccountID, password)
	_, _ = loginUser(t, refereePhone, password)

	t.Run("check_referral_summary_empty", func(t *testing.T) {
		summaryResp := getReferralSummary(t, referrerToken, referrerID)
		t.Logf("referrer summary before binding: %v", summaryResp)
	})

	t.Logf("registered referrer=%s referee=%s", referrerID, refereeID)
}

func TestOrderFlow(t *testing.T) {
	phone := fmt.Sprintf("136%08d", time.Now().UnixNano()%100000000)
	accountID := fmt.Sprintf("ordertest%d", time.Now().UnixNano())
	password := "OrderTest123!"

	registerUser(t, phone, accountID, password)
	loginBody, token := loginUser(t, phone, password)
	if token == "" {
		t.Fatal("login failed")
	}
	userIDFloat, _ := loginBody["user_id"].(float64)

	t.Run("create_order", func(t *testing.T) {
		orderResp := createOrder(t, token, int64(userIDFloat), "subscription", "Pro Monthly", 29.90)
		t.Logf("order created: %v", orderResp)
		assertKeyExists(t, orderResp, "data")
	})

	var orderID string
	t.Run("create_and_get_order", func(t *testing.T) {
		orderResp := createOrder(t, token, int64(userIDFloat), "subscription", "Pro Monthly", 29.90)
		data, ok := orderResp["data"].(map[string]interface{})
		if !ok {
			t.Fatal("order response missing data field")
		}
		orderID = fmt.Sprintf("%v", data["id"])
		t.Logf("created order ID: %s", orderID)

		getResp := getOrder(t, token, orderID)
		orderData, ok := getResp["data"].(map[string]interface{})
		if !ok {
			t.Fatal("get order response missing data field")
		}
		if orderData["status"] != "pending" {
			t.Errorf("expected pending status, got %v", orderData["status"])
		}
	})

	t.Run("update_order_to_paid", func(t *testing.T) {
		if orderID == "" {
			t.Skip("no order ID from previous step")
		}
		updateResp := updateOrderStatus(t, token, orderID, "paid")
		t.Logf("order updated to paid: %v", updateResp)

		getResp := getOrder(t, token, orderID)
		orderData, ok := getResp["data"].(map[string]interface{})
		if !ok {
			t.Fatal("get order response missing data field")
		}
		if orderData["status"] != "paid" {
			t.Errorf("expected paid status, got %v", orderData["status"])
		}
	})

	t.Run("list_orders", func(t *testing.T) {
		listResp := listOrders(t, token, int64(userIDFloat))
		assertKeyExists(t, listResp, "data")
		t.Logf("orders list: %v", listResp)
	})
}
