package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"account-center/sms-email-service/internal/handler"
	"account-center/sms-email-service/internal/provider"
	"account-center/sms-email-service/internal/service"
	"account-center/sms-email-service/pkg/circuitbreaker"
)

func main() {
	aliyunProvider := provider.NewAliyunProvider(
		getEnv("ALIYUN_APP_ID", ""),
		getEnv("ALIYUN_APP_SECRET", ""),
		getEnv("ALIYUN_SIGN_NAME", "AccountCenter"),
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

	aliyunCB := circuitbreaker.New(5, 30*time.Second)
	tencentCB := circuitbreaker.New(5, 30*time.Second)
	chinaTelecomCB := circuitbreaker.New(5, 30*time.Second)

	smsService := service.NewSMSService(
		[]provider.SMSProvider{aliyunProvider, tencentProvider, chinaTelecomProvider},
		[]*circuitbreaker.CircuitBreaker{aliyunCB, tencentCB, chinaTelecomCB},
	)
	smsHandler := handler.NewSMSHandler(smsService)

	r := gin.Default()

	smsGroup := r.Group("/api/v1/sms")
	{
		smsGroup.POST("/send", smsHandler.SendSMS)
		smsGroup.GET("/providers/status", smsHandler.GetProviderStatus)
	}

	port := getEnv("PORT", "8083")
	log.Printf("SMS service starting on :%s", port)

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

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}