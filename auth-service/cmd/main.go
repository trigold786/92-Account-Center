package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"account-center/account-service/internal/repository"
	"account-center/auth-service/internal/handler"
	"account-center/auth-service/internal/service"
	"account-center/auth-service/pkg/jwt"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "account_center")

	dsn := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser + " password=" + dbPassword + " dbname=" + dbName + " sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Connected to database")

	userRepo := repository.NewUserRepository(db)
	jwtManager := jwt.NewJWTManager(
		getEnv("JWT_ACCESS_SECRET", "access-secret-key-change-in-production"),
		getEnv("JWT_REFRESH_SECRET", "refresh-secret-key-change-in-production"),
		15*60*1000000000,
		7*24*60*60*1000000000,
	)
	authService := service.NewAuthService(userRepo, jwtManager)
	loginHandler := handler.NewLoginHandler(authService)

	r := gin.Default()

	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/login", loginHandler.Login)
		authGroup.POST("/login/send-sms-code", loginHandler.SendSMSCode)
		authGroup.POST("/login/send-email-otp", loginHandler.SendEmailOTP)
	}

	port := getEnv("PORT", "8082")
	log.Printf("Auth service starting on :%s", port)

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