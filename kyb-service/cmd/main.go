package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/trigold786/92-Account-Center/kyb-service/internal/handler"
	"github.com/trigold786/92-Account-Center/kyb-service/internal/repository"
	"github.com/trigold786/92-Account-Center/kyb-service/internal/service"
	"github.com/trigold786/92-Account-Center/kyb-service/pkg/crypto"
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

	encryptKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	entRepo := repository.NewEnterpriseRepository(db)
	kybService := service.NewKYBService(entRepo, encryptKey)
	kybHandler := handler.NewKYBHandler(kybService)

	r := gin.Default()

	kybGroup := r.Group("/api/v1/kyb")
	{
		kybGroup.POST("/submit", kybHandler.SubmitEnterprise)
		kybGroup.POST("/micro-payment/initiate", kybHandler.InitiateMicroPayment)
		kybGroup.POST("/micro-payment/verify", kybHandler.VerifyMicroPayment)
		kybGroup.POST("/face-verify", kybHandler.SubmitFaceVerification)
		kybGroup.GET("/status/:enterprise_id", kybHandler.GetEnterpriseStatus)
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30304")
	log.Printf("KYB service starting on :%s", port)

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
