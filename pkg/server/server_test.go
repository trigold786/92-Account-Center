package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestStandardServer(t *testing.T) {
	cfg := ServerConfig{
		Port:        "0",
		ServiceName: "test-service",
	}

	srv := NewStandardServer(cfg)
	srv.SetupMetrics()
	srv.SetupHealth(nil)

	srv.Engine().GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	req, _ = http.NewRequest("GET", "/health", nil)
	w = httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("health: expected 200, got %d", w.Code)
	}

	req, _ = http.NewRequest("GET", "/metrics", nil)
	w = httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("metrics: expected 200, got %d", w.Code)
	}
}

func TestGracefulShutdown(t *testing.T) {
	cfg := ServerConfig{
		Port:            "0",
		ServiceName:     "test-shutdown",
		ShutdownTimeout: 2 * time.Second,
	}

	srv := NewStandardServer(cfg)
	srv.SetupHealth(nil)

	shutdownCalled := make(chan struct{})
	srv.Engine().GET("/shutdown-test", func(c *gin.Context) {
		go func() {
			srv.Shutdown()
			close(shutdownCalled)
		}()
		c.JSON(http.StatusOK, gin.H{"shutting_down": true})
	})

	req, _ := http.NewRequest("GET", "/shutdown-test", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	select {
	case <-shutdownCalled:
	case <-time.After(5 * time.Second):
		t.Fatal("shutdown timed out")
	}
}

func TestHealthCheckFailure(t *testing.T) {
	cfg := ServerConfig{
		Port:        "0",
		ServiceName: "test-health",
	}

	srv := NewStandardServer(cfg)
	srv.SetupHealth(func(_ context.Context) error {
		return context.DeadlineExceeded
	})

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	srv.Engine().ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}
