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
	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/handler"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/repository"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/service"
	"github.com/trigold786/92-Account-Center/compliance-service/pkg/crypto"
	"github.com/trigold786/92-Account-Center/compliance-service/pkg/mq"
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

	redisURL := getEnv("REDIS_URL", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
	defer rdb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	riskRepo := repository.NewRiskRepository(db)
	geoService := service.NewGeoService()
	riskService := service.NewRiskService(riskRepo, geoService)
	riskHandler := handler.NewRiskHandler(riskService)

	auditRepo := repository.NewAuditRepository(db)
	auditService := service.NewAuditService(auditRepo)
	auditHandler := handler.NewAuditHandler(auditService)

	encryptKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("Failed to generate encryption key: %v", err)
	}

	entRepo := repository.NewEnterpriseRepository(db)
	kybService := service.NewKYBService(entRepo, encryptKey)
	kybHandler := handler.NewKYBHandler(kybService)

	r := gin.Default()

	riskHandler.RegisterRoutes(r)

	auditGroup := r.Group("/api/v1/audit")
	{
		auditGroup.POST("/logs", auditHandler.RecordLog)
		auditGroup.POST("/logs/batch", auditHandler.RecordBatch)
		auditGroup.GET("/logs/user/:user_id", auditHandler.GetLogsByUser)
		auditGroup.GET("/logs", auditHandler.GetLogsByTimeRange)
		auditGroup.GET("/logs/:log_id/verify", auditHandler.VerifyLogIntegrity)
		auditGroup.POST("/logs/cleanup", auditHandler.CleanupOldLogs)
	}

	kybGroup := r.Group("/api/v1/kyb")
	{
		kybGroup.POST("/submit", kybHandler.SubmitEnterprise)
		kybGroup.POST("/micro-payment/initiate", kybHandler.InitiateMicroPayment)
		kybGroup.POST("/micro-payment/verify", kybHandler.VerifyMicroPayment)
		kybGroup.POST("/face-verify", kybHandler.SubmitFaceVerification)
		kybGroup.GET("/status/:enterprise_id", kybHandler.GetEnterpriseStatus)
	}

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	mqType := getEnv("AUDIT_MQ_TYPE", "redis")

	var messageQueue mq.MessageQueue

	switch mqType {
	case "kafka":
		brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ",")
		topic := getEnv("KAFKA_AUDIT_TOPIC", "audit-logs")
		groupID := getEnv("KAFKA_GROUP_ID", "compliance-service")

		kafkaMQ, err := mq.NewKafkaMQ(brokers, topic, groupID)
		if err != nil {
			log.Printf("Warning: Failed to create Kafka MQ: %v (audit logs will be synchronous only)", err)
		} else {
			messageQueue = kafkaMQ
		}

	case "redis":
		streamKey := getEnv("REDIS_STREAM_KEY", "audit-logs")
		groupName := getEnv("REDIS_CONSUMER_GROUP", "compliance-service")
		consumerID := getEnv("REDIS_CONSUMER_ID", "compliance-service-1")

		redisPassword := getEnv("REDIS_PASSWORD", "")
		redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

		redisMQ := mq.NewRedisStreamsMQ(redisURL, redisPassword, redisDB, streamKey, groupName, consumerID)
		messageQueue = redisMQ

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

	port := getEnv("PORT", "30313")
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

	log.Printf("Compliance Service starting on :%s", port)
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
