package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/trigold786/92-Account-Center/audit-log-service/internal/handler"
	"github.com/trigold786/92-Account-Center/audit-log-service/internal/repository"
	"github.com/trigold786/92-Account-Center/audit-log-service/internal/service"
	"github.com/trigold786/92-Account-Center/audit-log-service/pkg/kafka"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnv("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "audit_db")

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

	auditRepo := repository.NewAuditRepository(db)
	auditService := service.NewAuditService(auditRepo)
	auditHandler := handler.NewAuditHandler(auditService)

	r := gin.Default()

	auditGroup := r.Group("/api/v1/audit")
	{
		auditGroup.POST("/logs", auditHandler.RecordLog)
		auditGroup.POST("/logs/batch", auditHandler.RecordBatch)
		auditGroup.GET("/logs/user/:user_id", auditHandler.GetLogsByUser)
		auditGroup.GET("/logs", auditHandler.GetLogsByTimeRange)
		auditGroup.GET("/logs/:log_id/verify", auditHandler.VerifyLogIntegrity)
		auditGroup.POST("/logs/cleanup", auditHandler.CleanupOldLogs)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaBrokers := []string{getEnv("KAFKA_BROKERS", "localhost:9092")}
	kafkaTopic := getEnv("KAFKA_AUDIT_TOPIC", "audit-logs")
	kafkaGroupID := getEnv("KAFKA_GROUP_ID", "audit-log-service")

	consumer, err := kafka.NewAuditLogConsumer(kafkaBrokers, kafkaGroupID, kafkaTopic, auditService)
	if err != nil {
		log.Printf("Warning: Failed to create Kafka consumer: %v (continuing without Kafka)", err)
	} else {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("Warning: Failed to start Kafka consumer: %v", err)
		} else {
			defer consumer.Close()
		}
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		os.Exit(0)
	}()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30305")
	log.Printf("Audit log service starting on :%s", port)
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
