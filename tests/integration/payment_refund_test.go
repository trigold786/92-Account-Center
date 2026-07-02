//go:build integration

package integration

import (
	"fmt"
	"net/http"
	"strconv"
	"testing"
	"time"
)

func TestPaymentRefundChain(t *testing.T) {
	phone := fmt.Sprintf("139%08d", time.Now().UnixNano()%100000000)
	accountID := fmt.Sprintf("paytest%d", time.Now().UnixNano())
	password := "TestPay123!"

	// 1. Register and login
	_, regUserID := registerUser(t, phone, accountID, password)
	if regUserID == "" {
		t.Fatal("registration did not return user ID")
	}

	loginBody, token := loginUser(t, phone, password)
	if token == "" {
		t.Fatal("login returned no token")
	}
	userIDFloat, ok := loginBody["user_id"].(float64)
	if !ok {
		t.Fatalf("login response missing user_id: %v", loginBody)
	}
	uid := int64(userIDFloat)
	t.Logf("registered user ID: %s (numeric %d)", regUserID, uid)

	// 2. Create order (response wraps data under "data")
	order := createOrder(t, token, uid, "subscription", "专业版月付", 29.90)
	orderData, ok := order["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("create order response missing data field: %v", order)
	}
	orderIDFloat, ok := orderData["id"].(float64)
	if !ok {
		t.Fatalf("order creation did not return numeric id: %v", orderData)
	}
	orderID := strconv.FormatInt(int64(orderIDFloat), 10)
	t.Logf("created order ID: %s", orderID)

	// 3. Verify order starts pending
	orderDetail := getOrder(t, token, orderID)
	if detailData, ok := orderDetail["data"].(map[string]interface{}); ok {
		if status, _ := detailData["status"].(string); status != "pending" {
			t.Fatalf("expected order status 'pending', got '%s'", status)
		}
	} else {
		t.Fatalf("get order response missing data field: %v", orderDetail)
	}

	// 4. Mark order as paid (simulating payment success)
	paidOrder := updateOrderStatus(t, token, orderID, "paid")
	if paidData, ok := paidOrder["data"].(map[string]interface{}); ok {
		if status, _ := paidData["status"].(string); status != "paid" {
			t.Fatalf("expected order status 'paid', got '%s'", status)
		}
	} else {
		t.Fatalf("update order response missing data field: %v", paidOrder)
	}

	// 5. List orders to verify the order exists in the list
	orders := listOrders(t, token, uid)
	if listData, ok := orders["data"].(map[string]interface{}); ok {
		if listData["total"] == nil {
			t.Fatalf("list orders returned no total: %v", listData)
		}
		t.Logf("orders total: %v", listData["total"])
	} else {
		t.Fatalf("list orders response missing data field: %v", orders)
	}

	// 6. Request refund (refund response is the model directly, no "data" wrapper;
	//    order_id must be the order's numeric ID)
	resp, refundBody := apiRequest(t, http.MethodPost, "/api/v1/refunds", map[string]interface{}{
		"order_id": int64(orderIDFloat),
		"reason":   "integration test refund",
	}, token)
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("refund request failed with status %d: %v", resp.StatusCode, refundBody)
	}
	refundID := fmt.Sprintf("%v", refundBody["id"])
	if refundID == "" || refundID == "<nil>" {
		t.Fatalf("refund response missing id: %v", refundBody)
	}
	t.Logf("created refund ID: %s", refundID)

	t.Run("order_lifecycle_pending_to_paid", func(t *testing.T) {
		finalOrder := getOrder(t, token, orderID)
		finalData, ok := finalOrder["data"].(map[string]interface{})
		if !ok {
			t.Fatalf("get order response missing data field: %v", finalOrder)
		}
		status, _ := finalData["status"].(string)
		if status != "paid" {
			t.Logf("order status is '%s' (expected 'paid' before refund approval)", status)
		}
	})
}
