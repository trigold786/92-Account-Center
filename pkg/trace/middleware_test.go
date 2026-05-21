package trace

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

func TestTracingMiddleware(t *testing.T) {
	shutdown, _ := InitProvider("test-service", "test", "")
	defer shutdown()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware("test-service"))
	r.GET("/test", func(c *gin.Context) {
		_, span := otel.Tracer("test").Start(c.Request.Context(), "handler")
		defer span.End()
		c.Status(200)
	})
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestTracingWithIncomingW3CHeader(t *testing.T) {
	shutdown, _ := InitProvider("test-service", "test", "")
	defer shutdown()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(Middleware("test-service"))
	r.GET("/test", func(c *gin.Context) {
		_, sc := otel.Tracer("test").Start(c.Request.Context(), "handler")
		traceID := sc.SpanContext().TraceID().String()
		if traceID == "" {
			t.Error("expected trace ID in context")
		}
		sc.End()
		c.Status(200)
	})
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("traceparent", "00-0af7651916cd43dd8448eb211c80319c-b7ad6b7169203331-01")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
