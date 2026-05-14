package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/notification-service/internal/handler"
	"github.com/trigold786/92-Account-Center/notification-service/internal/provider"
	"github.com/trigold786/92-Account-Center/notification-service/internal/service"
	"github.com/trigold786/92-Account-Center/notification-service/pkg/circuitbreaker"
)

var requestCount uint64

func main() {
	redisAddr := getEnv("REDIS_URL", "localhost:6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	service.SetRedisClient(redisClient)

	aliyunProvider := provider.NewAliyunProvider(
		getEnv("ALIYUN_ACCESS_KEY_ID", ""),
		getEnv("ALIYUN_ACCESS_KEY_SECRET", ""),
		getEnv("ALIYUN_SIGN_NAME", "速通互联验证码"),
	)
	tencentProvider := provider.NewTencentProvider(
		getEnv("TENCENT_APP_ID", ""),
		getEnv("TENCENT_APP_SECRET", ""),
		getEnv("TENCENT_SIGN_NAME", "AccountCenter"),
	)
	chinaTelecomProvider := provider.NewChinaTelecomProvider(
		getEnv("CHINATELECOM_APP_ID", ""),
		getEnv("CHINATELECOM_APP_SECRET", ""),
		getEnv("CHINATELECOM_SIGN_NAME", "AccountCenter"),
	)

	smsProviders := []provider.SMSProvider{aliyunProvider, tencentProvider, chinaTelecomProvider}
	circuitBreakers := []*circuitbreaker.CircuitBreaker{
		circuitbreaker.New(5, 30*time.Second),
		circuitbreaker.New(5, 30*time.Second),
		circuitbreaker.New(5, 30*time.Second),
	}

	smsService := service.NewSMSService(smsProviders, circuitBreakers)
	smsHandler := handler.NewSMSHandler(smsService)

	simpleSMTP := provider.NewSimpleSMTPProvider(
		getEnv("SMTP_HOST", "smtp.163.com"),
		getEnv("SMTP_PORT", "465"),
		getEnv("SMTP_USERNAME", ""),
		getEnv("SMTP_PASSWORD", ""),
		getEnv("SMTP_FROM", ""),
	)
	verificationEmailService := service.NewSimpleEmailService(simpleSMTP)
	verificationEmailHandler := handler.NewVerificationEmailHandler(verificationEmailService)

	emailProviderType := getEnv("EMAIL_PROVIDER", "smtp")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	fromAddress := getEnv("FROM_ADDRESS", "noreply@accountcenter.com")

	var emailProvider provider.EmailProvider

	switch emailProviderType {
	case "sendgrid":
		emailProvider = provider.NewSendGridProvider(getEnv("SENDGRID_API_KEY", ""), fromAddress)
	case "aws_ses":
		emailProvider = provider.NewSESProvider(
			getEnv("AWS_REGION", "us-east-1"),
			getEnv("AWS_ACCESS_KEY", ""),
			getEnv("AWS_SECRET_KEY", ""),
			fromAddress,
		)
	case "smtp":
		emailProvider = provider.NewSMTPProvider(
			getEnv("SMTP_HOST", "localhost"),
			getEnv("SMTP_PORT", "587"),
			getEnv("SMTP_USERNAME", ""),
			getEnv("SMTP_PASSWORD", ""),
			fromAddress,
		)
	default:
		log.Fatalf("Unsupported email provider: %s", emailProviderType)
	}
	log.Printf("Using email provider: %s", emailProvider.Name())

	otpEmailService := service.NewOTPEmailService(redisClient, emailProvider, jwtSecret, fromAddress)
	otpEmailHandler := handler.NewOTPEmailHandler(otpEmailService)

	pushService := service.NewPushService(redisClient)
	pushHandler := handler.NewPushHandler(pushService)

	r := gin.Default()

	r.Use(func(c *gin.Context) {
		atomic.AddUint64(&requestCount, 1)
		c.Next()
	})

	r.GET("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"notification-service\"} %d\n", atomic.LoadUint64(&requestCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"notification-service\"} %d\n", runtime.NumGoroutine())
		c.Data(http.StatusOK, "text/plain; version=0.0.4", buf.Bytes())
	})

	r.Any("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	smsGroup := r.Group("/api/v1/sms")
	{
		smsGroup.POST("/send", smsHandler.SendSMS)
		smsGroup.POST("/verify", smsHandler.VerifyCode)
		smsGroup.GET("/providers/status", smsHandler.GetProviderStatus)
	}

	emailGroup := r.Group("/api/v1/email")
	{
		emailGroup.POST("/verify", verificationEmailHandler.VerifyCode)
		emailGroup.POST("/otp/send", otpEmailHandler.SendOTP)
		emailGroup.POST("/otp/verify", otpEmailHandler.VerifyOTP)
		emailGroup.POST("/magic-link/send", otpEmailHandler.SendMagicLink)
		emailGroup.GET("/magic-link/verify", otpEmailHandler.VerifyMagicLink)
		emailGroup.POST("/send", otpEmailHandler.SendEmail)
	}

	pushGroup := r.Group("/api/v1/push")
	{
		pushGroup.POST("/send", pushHandler.SendPush)
		pushGroup.POST("/device/register", pushHandler.RegisterDevice)
		pushGroup.GET("/user/:user_id/devices", pushHandler.GetUserDevices)
	}

	port := getEnv("PORT", "30311")

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

	log.Printf("Notification service starting on :%s", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
