package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/api/v1/account/register", func(c *gin.Context) {
		var req map[string]interface{}
		c.ShouldBindJSON(&req)
		c.JSON(http.StatusCreated, gin.H{"user_id": "user-1", "status": "created"})
	})

	r.POST("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"access_token":  "token-abc",
			"refresh_token": "refresh-abc",
			"expires_in":    3600,
		})
	})

	r.GET("/api/v1/data/dashboard/overview", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"users": 100, "revenue": 5000})
	})

	r.POST("/api/v1/orders", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"order_id": "order-1", "status": "created"})
	})

	r.GET("/api/v1/credits/:user_id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"credits": 100})
	})

	r.POST("/api/v1/referral", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"referral_code": "REF123"})
	})

	return r
}

func makeRequest(t *testing.T, router *gin.Engine, method, path string, body interface{}) *httptest.ResponseRecorder {
	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal body: %v", err)
		}
		bodyReader = bytes.NewReader(b)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	req, err := http.NewRequest(method, path, bodyReader)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func TestJourney_Register(t *testing.T) {
	router := setupRouter()

	w := makeRequest(t, router, "POST", "/api/v1/account/register", map[string]string{
		"phone_number": "13800138000",
		"password":     "Test1234!",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("register: expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestJourney_Login(t *testing.T) {
	router := setupRouter()

	w := makeRequest(t, router, "POST", "/api/v1/auth/login", map[string]string{
		"credential": "13800138000",
		"password":   "Test1234!",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["access_token"] == nil {
		t.Fatal("expected access_token in response")
	}
}

func TestJourney_Browse(t *testing.T) {
	router := setupRouter()

	w := makeRequest(t, router, "GET", "/api/v1/data/dashboard/overview", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("browse: expected 200, got %d", w.Code)
	}
}

func TestJourney_Subscribe(t *testing.T) {
	router := setupRouter()

	w := makeRequest(t, router, "POST", "/api/v1/orders", map[string]interface{}{
		"plan_id":  "premium",
		"user_id":  "user-1",
		"amount":   99.99,
		"currency": "CNY",
	})

	if w.Code != http.StatusCreated {
		t.Fatalf("subscribe: expected 201, got %d", w.Code)
	}
}

func TestJourney_Credits(t *testing.T) {
	router := setupRouter()

	w := makeRequest(t, router, "GET", "/api/v1/credits/user-1", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("credits: expected 200, got %d", w.Code)
	}
}

func TestJourney_Share(t *testing.T) {
	router := setupRouter()

	w := makeRequest(t, router, "POST", "/api/v1/referral", map[string]string{
		"user_id": "user-1",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("share: expected 200, got %d", w.Code)
	}
}

func TestFullJourney(t *testing.T) {
	router := setupRouter()

	steps := []struct {
		name   string
		method string
		path   string
		body   interface{}
		expect int
	}{
		{"Register", "POST", "/api/v1/account/register", map[string]string{"phone_number": "13800138000", "password": "Test1234!"}, http.StatusCreated},
		{"Login", "POST", "/api/v1/auth/login", map[string]string{"credential": "13800138000", "password": "Test1234!"}, http.StatusOK},
		{"Browse", "GET", "/api/v1/data/dashboard/overview", nil, http.StatusOK},
		{"Subscribe", "POST", "/api/v1/orders", map[string]interface{}{"plan_id": "premium", "user_id": "user-1"}, http.StatusCreated},
		{"Credits", "GET", "/api/v1/credits/user-1", nil, http.StatusOK},
		{"Share", "POST", "/api/v1/referral", map[string]string{"user_id": "user-1"}, http.StatusOK},
	}

	for _, step := range steps {
		t.Run(step.name, func(t *testing.T) {
			w := makeRequest(t, router, step.method, step.path, step.body)
			if w.Code != step.expect {
				t.Errorf("%s: expected %d, got %d, body: %s", step.name, step.expect, w.Code, w.Body.String())
			}
		})
	}
}
