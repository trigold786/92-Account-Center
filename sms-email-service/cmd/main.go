package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/sms-email-service/internal/handler"
	"github.com/trigold786/92-Account-Center/sms-email-service/internal/provider"
	"github.com/trigold786/92-Account-Center/sms-email-service/internal/service"
	"github.com/trigold786/92-Account-Center/sms-email-service/pkg/circuitbreaker"
)

func main() {
	redisClient := redis.NewClient(&redis.Options{
		Addr: getEnv("REDIS_ADDR", "localhost:6379"),
	})

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

	providers := []provider.SMSProvider{aliyunProvider, tencentProvider, chinaTelecomProvider}
	circuitBreakers := []*circuitbreaker.CircuitBreaker{
		circuitbreaker.New(5, 30*time.Second),
		circuitbreaker.New(5, 30*time.Second),
		circuitbreaker.New(5, 30*time.Second),
	}

	service.SetRedisClient(redisClient)
	smsService := service.NewSMSService(providers, circuitBreakers)
	smsHandler := handler.NewSMSHandler(smsService)

	smtpProvider := provider.NewSMTPProvider(
		getEnv("SMTP_HOST", "smtp.163.com"),
		getEnv("SMTP_PORT", "465"),
		getEnv("SMTP_USERNAME", "trigoldsun@163.com"),
		getEnv("SMTP_PASSWORD", ""),
		getEnv("SMTP_FROM", "trigoldsun@163.com"),
	)
	emailService := service.NewEmailService(smtpProvider)
	emailHandler := handler.NewEmailHandler(emailService)

	r := gin.Default()

	smsGroup := r.Group("/api/v1/sms")
	{
		smsGroup.POST("/send", smsHandler.SendSMS)
		smsGroup.POST("/verify", smsHandler.VerifyCode)
		smsGroup.GET("/providers/status", smsHandler.GetProviderStatus)
	}

	emailGroup := r.Group("/api/v1/email")
	{
		emailGroup.POST("/send", emailHandler.SendVerificationCode)
		emailGroup.POST("/verify", emailHandler.VerifyCode)
	}

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30303")
	log.Printf("SMS/Email service starting on :%s", port)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		redisClient.Close()
		os.Exit(0)
	}()

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
