package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestVersionExtraction(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		path          string
		expectVersion string
	}{
		{"/api/v1/account/users", "v1"},
		{"/api/v2/account/users", "v2"},
		{"/api/v1/auth/login", "v1"},
		{"/api/v2/data/dashboard", "v2"},
	}

	for _, tt := range tests {
		r := gin.New()
		r.Use(VersionMiddleware())

		var gotVersion string
		r.Any("/api/:version/*path", func(c *gin.Context) {
			if v, ok := c.Get("api_version"); ok {
				gotVersion = v.(string)
			}
			c.Status(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if gotVersion != tt.expectVersion {
			t.Errorf("path %s: expected version %s, got %s", tt.path, tt.expectVersion, gotVersion)
		}
	}
}

func TestDeprecatedWarning(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(VersionMiddleware())
	r.Any("/api/:version/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/account/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	warning := w.Header().Get("X-Deprecated-Version")
	if warning == "" {
		t.Fatal("expected X-Deprecated-Version header for v1")
	}
}

func TestVersionNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(VersionMiddleware())
	r.Any("/api/:version/*path", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	v, exists := w.Result().Header["X-Deprecated-Version"]
	_ = v
	_ = exists
}
