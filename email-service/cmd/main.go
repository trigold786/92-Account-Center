package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"account-center/email-service/internal/handler"
	"account-center/email-service/internal/model"
	"account-center/email-service/internal/provider"
	"account-center/email-service/internal/service"
)

func main() {
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnv("REDIS_DB", "0")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	emailProviderType := getEnv("EMAIL_PROVIDER", "smtp")
	jwtSecret := getEnv("JWT_SECRET", "your-secret-key-change-in-production")
	fromAddress := getEnv("FROM_ADDRESS", "noreply@accountcenter.com")

	var emailProvider provider.EmailProvider

	switch emailProviderType {
	case "sendgrid":
		sendgridAPIKey := getEnv("SENDGRID_API_KEY", "")
		emailProvider = provider.NewSendGridProvider(sendgridAPIKey, fromAddress)
	case "aws_ses":
		awsRegion := getEnv("AWS_REGION", "us-east-1")
		awsAccessKey := getEnv("AWS_ACCESS_KEY", "")
		awsSecretKey := getEnv("AWS_SECRET_KEY", "")
		emailProvider = provider.NewSESProvider(awsRegion, awsAccessKey, awsSecretKey, fromAddress)
	case "smtp":
		smtpHost := getEnv("SMTP_HOST", "localhost")
		smtpPort := getEnv("SMTP_PORT", "587")
		smtpUsername := getEnv("SMTP_USERNAME", "")
		smtpPassword := getEnv("SMTP_PASSWORD", "")
		emailProvider = provider.NewSMTPProvider(smtpHost, smtpPort, smtpUsername, smtpPassword, fromAddress)
	default:
		log.Fatalf("Unsupported email provider: %s", emailProviderType)
	}
	log.Printf("Using email provider: %s", emailProvider.Name())

	emailService := service.NewEmailService(redisClient, emailProvider, jwtSecret, fromAddress)
	emailHandler := handler.NewEmailHandler(emailService)

	r := gin.Default()

	emailGroup := r.Group("/email")
	{
		emailGroup.POST("/otp/send", emailHandler.SendOTP)
		emailGroup.POST("/otp/verify", emailHandler.VerifyOTP)
		emailGroup.POST("/magic-link/send", emailHandler.SendMagicLink)
		emailGroup.GET("/magic-link/verify", emailHandler.VerifyMagicLink)
		emailGroup.POST("/send", emailHandler.SendEmail)
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	port := getEnv("PORT", "8080")
	log.Printf("Email service starting on :%s", port)
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
