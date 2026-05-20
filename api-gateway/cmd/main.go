package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	"strings"
	"sync"
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

type tokenBucket struct {
	tokens  float64
	lastRef time.Time
	mu      sync.Mutex
}

type ipRateLimiter struct {
	buckets sync.Map
	rps     float64
}

func newIPRateLimiter(rps int) *ipRateLimiter {
	return &ipRateLimiter{rps: float64(rps)}
}

func (l *ipRateLimiter) getBucket(ip string) *tokenBucket {
	val, _ := l.buckets.LoadOrStore(ip, &tokenBucket{
		tokens:  l.rps,
		lastRef: time.Now(),
	})
	return val.(*tokenBucket)
}

func (l *ipRateLimiter) allow(ip string) bool {
	b := l.getBucket(ip)
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.lastRef).Seconds()
	b.lastRef = now
	b.tokens += elapsed * l.rps
	if b.tokens > l.rps {
		b.tokens = l.rps
	}
	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

func cacheControlMiddleware(maxAge int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
		}
		c.Next()
	}
}

func jwtAuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenStr := parts[1]
		segments := strings.Split(tokenStr, ".")
		if len(segments) != 3 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			return
		}

		signingInput := segments[0] + "." + segments[1]
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(signingInput))
		expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(segments[2]), []byte(expectedSig)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token signature"})
			return
		}

		payloadBytes, err := base64.RawURLEncoding.DecodeString(segments[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
			return
		}

		var claims map[string]interface{}
		if err := json.Unmarshal(payloadBytes, &claims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
				return
			}
		}

		userID, _ := claims["sub"].(string)
		if userID == "" {
			userID, _ = claims["user_id"].(string)
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

func rateLimitMiddleware(rps int) gin.HandlerFunc {
	limiter := newIPRateLimiter(rps)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}

func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func requestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Header("X-Request-ID", requestID)
		c.Next()
	}
}

var logger *slog.Logger

func init() { slog.SetDefault(logger) }

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

	r.Use(requestIDMiddleware())
	r.Use(corsMiddleware())
	r.Use(rateLimitMiddleware(svcCfg.RateLimitRPS))
	r.Use(cacheControlMiddleware(svcCfg.CacheMaxAge))
	r.Use(middleware.TimeoutMiddleware(svcCfg.GlobalRequestTimeoutSec))

	publicPaths := map[string]bool{
		"/api/v1/auth/login":              true,
		"/api/v1/auth/refresh":            true,
		"/api/v1/auth/biometric/login":    true,
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
		jwtAuthMiddleware(svcCfg.JWTSecret)(c)
	})

	r.Use(desensitizeMiddleware(svcCfg.MaxDesensitizeBodySize))

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

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		allowedOrigins := map[string]bool{
			"http://localhost:30317": true,
			"http://localhost:30316": true,
			getEnv("WEB_UI_ORIGIN", ""): true,
		}
		if allowedOrigins[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		} else {
			c.Header("Access-Control-Allow-Origin", "http://localhost:30317")
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
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

type responseCaptureWriter struct {
	gin.ResponseWriter
	body []byte
	code int
}

func (w *responseCaptureWriter) Status() int {
	return w.code
}

func (w *responseCaptureWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}

func (w *responseCaptureWriter) WriteHeader(code int) {
	w.code = code
}

var (
	gatewayRequestCount uint64
	durationSumNano     uint64
	durationCount       uint64
)

var phoneRegex = regexp.MustCompile(`"phone_number"\s*:\s*"(\d{3})\d{4}(\d{4})"`)
var emailRegex = regexp.MustCompile(`"email"\s*:\s*"([a-zA-Z0-9])[a-zA-Z0-9._%+\-]*@([^"]+)"`)
var ipAddrRegex = regexp.MustCompile(`"ip_address"\s*:\s*"(\d{1,3}\.)\d{1,3}\.\d{1,3}(\.\d{1,3})"`)

func desensitizeMiddleware(maxBodySize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/health" || path == "/metrics" || strings.HasPrefix(path, "/internal/") {
			c.Next()
			return
		}

		captureWriter := &responseCaptureWriter{ResponseWriter: c.Writer}
		c.Writer = captureWriter

		c.Next()

		status := captureWriter.code

		flush := func(data []byte) {
			captureWriter.ResponseWriter.WriteHeader(status)
			captureWriter.ResponseWriter.Write(data)
		}

		if status < 200 || status >= 300 {
			flush(captureWriter.body)
			return
		}

		contentType := c.Writer.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			flush(captureWriter.body)
			return
		}

		if int64(len(captureWriter.body)) > maxBodySize {
			flush(captureWriter.body)
			return
		}

		accountID, _ := c.Get("user_id")
		if accountIDStr, ok := accountID.(string); ok && strings.HasPrefix(accountIDStr, "admin_") {
			flush(captureWriter.body)
			return
		}

		body := string(captureWriter.body)
		masked := phoneRegex.ReplaceAllString(body, `"phone_number":"$1****$2"`)
		masked = emailRegex.ReplaceAllString(masked, `"email":"$1***@$2"`)
		masked = ipAddrRegex.ReplaceAllString(masked, `"ip_address":"$1*.*$2"`)

		if masked != body {
			c.Header("X-Desensitized", "true")
			captureWriter.ResponseWriter.Header().Del("Content-Length")
			captureWriter.ResponseWriter.WriteHeader(status)
			captureWriter.ResponseWriter.Write([]byte(masked))
		} else {
			flush(captureWriter.body)
		}
	}
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
