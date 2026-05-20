package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupPricingRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPricingHandler()
	r.GET("/api/v1/pricing", h.GetPricing)
	r.POST("/api/v1/pricing/calculate-discount", h.CalculateDiscount)
	return r
}

func TestPricing_GetPricing(t *testing.T) {
	r := setupPricingRouter()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/pricing", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	tiers, ok := resp["tiers"].([]interface{})
	if !ok {
		t.Fatal("tiers not found or not an array")
	}
	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}

	if creditRate, ok := resp["credit_rate"].(float64); !ok || creditRate != 0.01 {
		t.Fatalf("expected credit_rate 0.01, got %v", resp["credit_rate"])
	}
	if maxDiscount, ok := resp["max_credit_discount"].(float64); !ok || maxDiscount != 0.5 {
		t.Fatalf("expected max_credit_discount 0.5, got %v", resp["max_credit_discount"])
	}

	tierNames := []string{"基础版", "专业版", "企业版"}
	tierLevels := []float64{2, 3, 4}
	monthlyPrices := []float64{9.9, 29.9, 99.9}
	for i, tier := range tiers {
		tierMap := tier.(map[string]interface{})
		if tierMap["name"] != tierNames[i] {
			t.Errorf("tier %d: expected name %s, got %v", i, tierNames[i], tierMap["name"])
		}
		if tierMap["tier_level"] != tierLevels[i] {
			t.Errorf("tier %d: expected level %v, got %v", i, tierLevels[i], tierMap["tier_level"])
		}
		if tierMap["monthly_price"] != monthlyPrices[i] {
			t.Errorf("tier %d: expected monthly_price %v, got %v", i, monthlyPrices[i], tierMap["monthly_price"])
		}
		features, ok := tierMap["features"].([]interface{})
		if !ok || len(features) == 0 {
			t.Errorf("tier %d: features missing or empty", i)
		}
		entitlements, ok := tierMap["entitlements"].([]interface{})
		if !ok || len(entitlements) == 0 {
			t.Errorf("tier %d: entitlements missing or empty", i)
		}
	}
}

func TestPricing_CalculateDiscount_ValidInput(t *testing.T) {
	r := setupPricingRouter()

	body := `{"price": 29.9, "credits_used": 500}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/pricing/calculate-discount", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if resp["original_price"].(float64) != 29.9 {
		t.Errorf("expected original_price 29.9, got %v", resp["original_price"])
	}
	if resp["credits_used"].(float64) != 500 {
		t.Errorf("expected credits_used 500, got %v", resp["credits_used"])
	}
	if resp["final_price"].(float64) != 24.9 {
		t.Errorf("expected final_price 24.9, got %v", resp["final_price"])
	}
	if resp["credit_value"].(float64) != 5.0 {
		t.Errorf("expected credit_value 5.0, got %v", resp["credit_value"])
	}
}

func TestPricing_CalculateDiscount_ExceedsCap(t *testing.T) {
	r := setupPricingRouter()

	body := `{"price": 29.9, "credits_used": 5000}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/pricing/calculate-discount", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	maxDiscount := 29.9 * 0.5
	if resp["credit_value"].(float64) != maxDiscount {
		t.Errorf("expected credit_value capped at %v, got %v", maxDiscount, resp["credit_value"])
	}
	if resp["final_price"].(float64) != maxDiscount {
		t.Errorf("expected final_price %v, got %v", maxDiscount, resp["final_price"])
	}
	if resp["discount_percent"].(float64) != 50.0 {
		t.Errorf("expected discount_percent 50.0, got %v", resp["discount_percent"])
	}
}

func TestPricing_CalculateDiscount_ZeroCredits(t *testing.T) {
	r := setupPricingRouter()

	body := `{"price": 29.9, "credits_used": 0}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/pricing/calculate-discount", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["final_price"].(float64) != 29.9 {
		t.Errorf("expected final_price 29.9, got %v", resp["final_price"])
	}
	if resp["credit_value"].(float64) != 0 {
		t.Errorf("expected credit_value 0, got %v", resp["credit_value"])
	}
	if resp["discount_percent"].(float64) != 0 {
		t.Errorf("expected discount_percent 0, got %v", resp["discount_percent"])
	}
}

func TestPricing_CalculateDiscount_InvalidInput(t *testing.T) {
	r := setupPricingRouter()

	tests := []struct {
		name string
		body string
	}{
		{"negative price", `{"price": -10, "credits_used": 100}`},
		{"zero price", `{"price": 0, "credits_used": 100}`},
		{"negative credits", `{"price": 29.9, "credits_used": -5}`},
		{"missing price", `{"credits_used": 100}`},
		{"empty body", ``},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/v1/pricing/calculate-discount", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			if w.Code != http.StatusBadRequest {
				t.Errorf("%s: expected status 400, got %d", tt.name, w.Code)
			}
		})
	}
}
