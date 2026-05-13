package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/trigold786/92-Account-Center/account-service/internal/handler"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
	"github.com/trigold786/92-Account-Center/account-service/pkg/sms"
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
	smsClient := sms.NewClient(getEnv("SMS_SERVICE_URL", "http://localhost:8083"))
	userService := service.NewUserService(userRepo, smsClient)

	registerHandler := handler.NewRegisterHandler(userService)
	passwordHandler := handler.NewPasswordHandler(userService)
	deletionHandler := handler.NewDeletionHandler(userService)

	r := gin.Default()

	accountGroup := r.Group("/api/v1/account")
	{
		accountGroup.POST("/register", registerHandler.Register)

		accountGroup.POST("/password/send-verification-code", passwordHandler.SendVerificationCode)
		accountGroup.POST("/password/change", passwordHandler.ChangePassword)

		accountGroup.POST("/deletion/request", deletionHandler.RequestDeletion)
		accountGroup.POST("/deletion/cancel", deletionHandler.CancelDeletion)
		accountGroup.GET("/deletion/status", deletionHandler.GetDeletionStatus)
	}

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30301")
	log.Printf("Account service starting on :%s", port)

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
