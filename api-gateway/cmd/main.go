package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type GatewayConfig struct {
	AccountServiceURL string
	AuthServiceURL    string
	SMSServiceURL     string
}

func main() {
	config := GatewayConfig{
		AccountServiceURL: getEnv("ACCOUNT_SERVICE_URL", "http://localhost:8081"),
		AuthServiceURL:    getEnv("AUTH_SERVICE_URL", "http://localhost:8082"),
		SMSServiceURL:     getEnv("SMS_SERVICE_URL", "http://localhost:8083"),
	}

	r := gin.Default()

	r.Use(corsMiddleware())

	r.Any("/api/v1/account/*path", proxyHandler(config.AccountServiceURL))
	r.Any("/api/v1/auth/*path", proxyHandler(config.AuthServiceURL))
	r.Any("/api/v1/sms/*path", proxyHandler(config.SMSServiceURL))
	r.Any("/api/v1/kyb/*path", proxyHandler(getEnv("KYB_SERVICE_URL", "http://localhost:8084")))
	r.Any("/api/v1/audit/*path", proxyHandler(getEnv("AUDIT_SERVICE_URL", "http://localhost:8085")))
	r.Any("/api/v1/risk/*path", proxyHandler(getEnv("RISK_SERVICE_URL", "http://localhost:8086")))
	r.Any("/api/v1/session/*path", proxyHandler(getEnv("SESSION_SERVICE_URL", "http://localhost:8087")))
	r.Any("/api/v1/email/*path", proxyHandler(getEnv("EMAIL_SERVICE_URL", "http://localhost:8088")))
	r.Any("/api/v1/device/*path", proxyHandler(getEnv("DEVICE_SERVICE_URL", "http://localhost:8089")))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "8080")
	log.Printf("API Gateway starting on :%s", port)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	if err := r.Run(":" + port); err != nil {
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
		c.Request.URL.Path = c.Param("path")
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