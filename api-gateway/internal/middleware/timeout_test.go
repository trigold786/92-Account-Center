package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestTimeoutMiddleware_ExceedsTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TimeoutMiddleware(1))
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusGatewayTimeout {
		t.Errorf("expected 504, got %d", w.Code)
	}
}

func TestTimeoutMiddleware_WithinTimeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TimeoutMiddleware(5))
	r.GET("/fast", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/fast", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestTimeoutMiddleware_ResponseBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TimeoutMiddleware(1))
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.JSON(200, gin.H{"ok": true})
	})
	req := httptest.NewRequest("GET", "/slow", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 504 {
		t.Fatalf("expected 504, got %d", w.Code)
	}
	body := w.Body.String()
	if !strings.Contains(body, "request timed out") {
		t.Fatalf("expected body to contain 'request timed out', got: %s", body)
	}
}

func TestTimeoutMiddleware_AlreadyAborted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(TimeoutMiddleware(1))
	r.GET("/abort", func(c *gin.Context) {
		c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
		time.Sleep(2 * time.Second)
	})
	req := httptest.NewRequest("GET", "/abort", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 403 {
		t.Fatalf("expected 403 from handler abort, got %d", w.Code)
	}
}

func TestTimeoutMiddleware_HandlerObservesCancellation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cancelled := make(chan bool, 1)
	r := gin.New()
	r.Use(TimeoutMiddleware(1))
	r.GET("/cancel", func(c *gin.Context) {
		select {
		case <-c.Request.Context().Done():
			cancelled <- true
		case <-time.After(3 * time.Second):
			cancelled <- false
		}
	})
	req := httptest.NewRequest("GET", "/cancel", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 504 {
		t.Fatalf("expected 504, got %d", w.Code)
	}
	select {
	case wasCancelled := <-cancelled:
		if !wasCancelled {
			t.Fatal("expected handler to observe context cancellation")
		}
	case <-time.After(4 * time.Second):
		t.Fatal("handler did not complete within timeout")
	}
}
