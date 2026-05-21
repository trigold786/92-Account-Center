package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/api-gateway/internal/middleware"
	proxyutil "github.com/trigold786/92-Account-Center/api-gateway/internal/proxy"
	"github.com/trigold786/92-Account-Center/api-gateway/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/pkg/config"
	healthpkg "github.com/trigold786/92-Account-Center/pkg/health"
	"github.com/trigold786/92-Account-Center/pkg/logging"
)

var logger *slog.Logger

func init() {}

var (
	gatewayRequestCount uint64
	durationSumNano     uint64
	durationCount       uint64
)

func main() {
	logger = logging.NewLogger("api-gateway")
	configURL := getEnv("CONFIG_SERVICE_URL", "http://localhost:30315")
	configClient := config.NewClient(configURL)
	svcCfg, err := svcconfig.Load(configClient)
	if err != nil {
		logger.Warn("config-service unavailable, continuing with env/defaults", "error", err)
	}
	logger.Info("gateway config loaded successfully")

	compositeHealth := healthpkg.CompositeChecker{}

	r := gin.New()
	r.Use(gin.RecoveryWithWriter(os.Stderr, func(c *gin.Context, err any) {
		logger.Error("panic recovered", "error", fmt.Sprintf("%v", err))
	}))
	r.Use(logging.Middleware(logger))

	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		atomic.AddUint64(&gatewayRequestCount, 1)
		atomic.AddUint64(&durationSumNano, uint64(time.Since(start).Nanoseconds()))
		atomic.AddUint64(&durationCount, 1)
	})

	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.VersionMiddleware())
	r.Use(middleware.RateLimitMiddleware(svcCfg.RateLimitRPS))
	r.Use(cacheControlMiddleware(svcCfg.CacheMaxAge))
	r.Use(middleware.TimeoutMiddleware(svcCfg.GlobalRequestTimeoutSec))

	publicPaths := map[string]bool{
		"/api/v1/auth/login":              true,
		"/api/v1/auth/refresh":            true,
		"/api/v1/auth/biometric/login":    true,
		"/api/v1/auth/oauth/":             true,
		"/api/v1/auth/enterprise/":        true,
		"/api/v1/auth/guest":              true,
		"/api/v1/auth/guest/":             true,
		"/api/v1/account/register":        true,
		"/api/v1/qrcode/generate":         true,
		"/api/v1/qrcode/":                 true,
		"/api/v1/sms/send":                true,
		"/api/v1/sms/verify":              true,
		"/api/v1/email/otp/send":          true,
		"/api/v1/email/magic-link/":       true,
	}

	r.Use(func(c *gin.Context) {
		for prefix := range publicPaths {
			if len(c.Request.URL.Path) >= len(prefix) && c.Request.URL.Path[:len(prefix)] == prefix {
				c.Next()
				return
			}
		}
		if c.Request.URL.Path == "/health" || c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}
		middleware.JWTAuthMiddleware(svcCfg.JWTSecret)(c)
	})

	r.Use(middleware.DesensitizeMiddleware(svcCfg.MaxDesensitizeBodySize))
	r.Use(middleware.HMACVerifyMiddleware(svcCfg.JWTSecret))
	r.Use(middleware.SanitizeInputMiddleware())

	r.Any("/api/v1/account/*path", proxyHandler(svcCfg.AccountServiceURL, svcCfg))
	r.Any("/api/v1/entitlements/*path", proxyHandler(svcCfg.AccountServiceURL, svcCfg))
	r.Any("/api/v1/subscriptions/*path", proxyHandler(svcCfg.AccountServiceURL, svcCfg))
	r.Any("/api/v1/auth/*path", proxyHandler(svcCfg.AuthServiceURL, svcCfg))
	r.Any("/api/v1/session/*path", proxyHandler(svcCfg.AuthServiceURL, svcCfg))
	r.Any("/api/v1/device/*path", proxyHandler(svcCfg.AuthServiceURL, svcCfg))
	r.Any("/api/v1/qrcode/*path", proxyHandler(svcCfg.AuthServiceURL, svcCfg))
	r.Any("/api/v1/sms/*path", proxyHandler(svcCfg.NotificationServiceURL, svcCfg))
	r.Any("/api/v1/email/*path", proxyHandler(svcCfg.NotificationServiceURL, svcCfg))
	r.Any("/api/v1/push/*path", proxyHandler(svcCfg.NotificationServiceURL, svcCfg))
	r.Any("/api/v1/credits/*path", proxyHandler(svcCfg.CreditServiceURL, svcCfg))
	r.Any("/api/v1/referral/*path", proxyHandler(svcCfg.CreditServiceURL, svcCfg))
	r.Any("/api/v1/risk/*path", proxyHandler(svcCfg.ComplianceServiceURL, svcCfg))
	r.Any("/api/v1/audit/*path", proxyHandler(svcCfg.ComplianceServiceURL, svcCfg))
	r.Any("/api/v1/kyb/*path", proxyHandler(svcCfg.ComplianceServiceURL, svcCfg))
	r.Any("/api/v1/data/*path", proxyHandler(svcCfg.DataProductServiceURL, svcCfg))

	v2Routes := []struct {
		prefix  string
		target  string
	}{
		{"/api/v2/account", svcCfg.AccountServiceURL},
		{"/api/v2/entitlements", svcCfg.AccountServiceURL},
		{"/api/v2/subscriptions", svcCfg.AccountServiceURL},
		{"/api/v2/auth", svcCfg.AuthServiceURL},
		{"/api/v2/session", svcCfg.AuthServiceURL},
		{"/api/v2/device", svcCfg.AuthServiceURL},
		{"/api/v2/credits", svcCfg.CreditServiceURL},
		{"/api/v2/data", svcCfg.DataProductServiceURL},
	}
	for _, route := range v2Routes {
		r.Any(route.prefix+"/*path", v2ProxyHandler(route.target, svcCfg))
	}

	r.GET("/api/v1/security/pins", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"pins": []string{"sha256/AAAA...", "sha256/BBBB..."}})
	})

	r.Any("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"api-gateway\"} %d\n", atomic.LoadUint64(&gatewayRequestCount))
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_sum Total request duration in seconds\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_sum counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_sum{service=\"api-gateway\"} %f\n", time.Duration(atomic.LoadUint64(&durationSumNano)).Seconds())
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_count Total request count for duration\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_count counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_count{service=\"api-gateway\"} %d\n", atomic.LoadUint64(&durationCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"api-gateway\"} %d\n", runtime.NumGoroutine())
		c.Data(http.StatusOK, "text/plain; version=0.0.4", buf.Bytes())
	})

	r.Any("/health", func(c *gin.Context) {
		result := compositeHealth.Check(c.Request.Context())
		resp := healthpkg.BuildResponse(result.Checks)
		statusCode := 200
		if result.Status == healthpkg.StatusDown {
			statusCode = 503
		}
		c.JSON(statusCode, resp)
	})

	srv := &http.Server{
		Addr:    ":" + svcCfg.Port,
		Handler: r,
	}

	go func() {
		defer logging.RecoverGoroutine(logger, "shutdown")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), svcCfg.ShutdownTimeout)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	logger.Info("starting server", "port", svcCfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("failed to start server", "error", err.Error())
		os.Exit(1)
	}
}

func cacheControlMiddleware(maxAge int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
		}
		c.Next()
	}
}

func proxyHandler(target string, cfg *svcconfig.GatewayConfig) gin.HandlerFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		logger.Error("invalid target URL", "error", err.Error())
		os.Exit(1)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = proxyutil.NewTransport(cfg.ResponseHeaderTimeoutSec, cfg.IdleConnTimeoutSec)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("proxy error", "target", target, "error", err.Error())
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}

	return func(c *gin.Context) {
		c.Request.URL.Host = targetURL.Host
		c.Request.URL.Scheme = targetURL.Scheme
		c.Request.Host = targetURL.Host

		c.Header("X-Request-ID", c.GetHeader("X-Request-ID"))
		c.Header("X-Forwarded-For", c.ClientIP())

		proxy.ServeHTTP(wrapWriter(c), c.Request)
	}
}

func v2ProxyHandler(target string, cfg *svcconfig.GatewayConfig) gin.HandlerFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		logger.Error("invalid target URL", "error", err.Error())
		os.Exit(1)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.Transport = proxyutil.NewTransport(cfg.ResponseHeaderTimeoutSec, cfg.IdleConnTimeoutSec)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		logger.Error("proxy error", "target", target, "error", err.Error())
		http.Error(w, "Service unavailable", http.StatusServiceUnavailable)
	}

	return func(c *gin.Context) {
		c.Request.URL.Host = targetURL.Host
		c.Request.URL.Scheme = targetURL.Scheme
		c.Request.Host = targetURL.Host
		c.Request.Header.Set("X-API-Version", "2")

		rewrittenPath := strings.Replace(c.Request.URL.Path, "/api/v2/", "/api/v1/", 1)
		c.Request.URL.Path = rewrittenPath

		c.Header("X-Request-ID", c.GetHeader("X-Request-ID"))
		c.Header("X-Forwarded-For", c.ClientIP())
		c.Header("X-API-Version", "2")

		proxy.ServeHTTP(wrapWriter(c), c.Request)
	}
}

type responseWriter struct {
	gin.ResponseWriter
}

func (w *responseWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

func wrapWriter(c *gin.Context) http.ResponseWriter {
	return &responseWriter{ResponseWriter: c.Writer}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvSecret(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	logger.Warn("environment variable not set, using insecure default", "key", key)
	return defaultValue
}

func sanitizeInputMiddleware() gin.HandlerFunc {
	return middleware.SanitizeInputMiddleware()
}

func hmacVerifyMiddleware(secret string) gin.HandlerFunc {
	return middleware.HMACVerifyMiddleware(secret)
}

func desensitizeMiddleware(maxBodySize int64) gin.HandlerFunc {
	return middleware.DesensitizeMiddleware(maxBodySize)
}

func jwtAuthMiddleware(secret string) gin.HandlerFunc {
	return middleware.JWTAuthMiddleware(secret)
}

func rateLimitMiddleware(rps int) gin.HandlerFunc {
	return middleware.RateLimitMiddleware(rps)
}

func corsMiddleware() gin.HandlerFunc {
	return middleware.CORSMiddleware()
}

func requestIDMiddleware() gin.HandlerFunc {
	return middleware.RequestIDMiddleware()
}

func generateRequestID() string {
	return middleware.GenerateRequestID()
}

func sanitizeString(s string) string {
	return middleware.SanitizeString(s)
}
