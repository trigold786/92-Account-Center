package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"

	"github.com/trigold786/92-Account-Center/audit-log-service/internal/handler"
	"github.com/trigold786/92-Account-Center/audit-log-service/internal/repository"
	"github.com/trigold786/92-Account-Center/audit-log-service/internal/service"
	"github.com/trigold786/92-Account-Center/audit-log-service/pkg/mq"
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

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mqType := getEnv("AUDIT_MQ_TYPE", "redis")

	var messageQueue mq.MessageQueue

	switch mqType {
	case "kafka":
		brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
		topic := getEnv("KAFKA_AUDIT_TOPIC", "audit-logs")
		groupID := getEnv("KAFKA_GROUP_ID", "audit-log-service")

		kafkaMQ, err := mq.NewKafkaMQ(brokers, topic, groupID)
		if err != nil {
			log.Printf("Warning: Failed to create Kafka MQ: %v (audit logs will be synchronous only)", err)
		} else {
			messageQueue = kafkaMQ
		}

	case "redis":
		redisAddr := getEnv("REDIS_ADDR", "localhost:6379")
		redisPassword := getEnv("REDIS_PASSWORD", "")
		redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
		streamKey := getEnv("REDIS_STREAM_KEY", "audit-logs")
		groupName := getEnv("REDIS_CONSUMER_GROUP", "audit-log-service")
		consumerID := getEnv("REDIS_CONSUMER_ID", "w004-audit-1")

		messageQueue = mq.NewRedisStreamsMQ(redisAddr, redisPassword, redisDB, streamKey, groupName, consumerID)

	default:
		log.Printf("Unknown AUDIT_MQ_TYPE=%s, falling back to synchronous-only mode", mqType)
	}

	if messageQueue != nil {
		defer messageQueue.Close()
		if err := messageQueue.StartConsumer(ctx, auditService); err != nil {
			log.Printf("Warning: Failed to start MQ consumer: %v (audit logs will be synchronous only)", err)
		} else {
			log.Printf("MQ consumer started: type=%s", mqType)
		}
	} else {
		log.Printf("No message queue configured, audit logs are synchronous only")
	}

	port := getEnv("PORT", "30305")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server forced shutdown: %v", err)
		}
	}()

	log.Printf("Audit log service starting on :%s", port)
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
