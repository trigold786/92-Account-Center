package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"context"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type GatewayConfig struct {
	AccountServiceURL string
	AuthServiceURL    string
	SMSServiceURL     string
}

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

func main() {
	config := GatewayConfig{
		AccountServiceURL: getEnv("ACCOUNT_SERVICE_URL", "http://localhost:30301"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://localhost:30302"),
		SMSServiceURL:     getEnv("SMS_SERVICE_URL", "http://localhost:30303"),
	}

	r := gin.Default()

	r.Use(requestIDMiddleware())
	r.Use(corsMiddleware())
	r.Use(rateLimitMiddleware(100))

	publicPaths := map[string]bool{
		"/api/v1/auth/login":         true,
		"/api/v1/auth/refresh":       true,
		"/api/v1/account/register":   true,
		"/api/v1/sms/":               true,
		"/api/v1/email/otp/send":     true,
		"/api/v1/email/magic-link/":  true,
		"/metrics":                   true,
	}

	r.Use(func(c *gin.Context) {
		for prefix := range publicPaths {
			if len(c.Request.URL.Path) >= len(prefix) && c.Request.URL.Path[:len(prefix)] == prefix {
				c.Next()
				return
			}
		}
		if c.Request.URL.Path == "/health" {
			c.Next()
			return
		}
		jwtAuthMiddleware(getEnv("JWT_SECRET", "default-secret"))(c)
	})

	r.Any("/api/v1/account/*path", proxyHandler(config.AccountServiceURL))
	r.Any("/api/v1/auth/*path", proxyHandler(config.AuthServiceURL))
	r.Any("/api/v1/sms/*path", proxyHandler(config.SMSServiceURL))
	r.Any("/api/v1/kyb/*path", proxyHandler(getEnv("KYB_SERVICE_URL", "http://localhost:30304")))
	r.Any("/api/v1/audit/*path", proxyHandler(getEnv("AUDIT_SERVICE_URL", "http://localhost:30305")))
	r.Any("/api/v1/risk/*path", proxyHandler(getEnv("RISK_SERVICE_URL", "http://localhost:30306")))
	r.Any("/api/v1/session/*path", proxyHandler(getEnv("SESSION_SERVICE_URL", "http://localhost:30307")))
	r.Any("/api/v1/email/*path", proxyHandler(getEnv("EMAIL_SERVICE_URL", "http://localhost:30308")))
	r.Any("/api/v1/device/*path", proxyHandler(getEnv("DEVICE_SERVICE_URL", "http://localhost:30309")))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	port := getEnv("PORT", "30300")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("API Gateway starting on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func proxyHandler(target string) gin.HandlerFunc {
	targetURL, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid target URL: %v", err)
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("Proxy error: %v", err)
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
		c.Header("Access-Control-Allow-Origin", "*")
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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}