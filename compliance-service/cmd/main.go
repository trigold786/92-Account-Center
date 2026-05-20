package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/handler"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/repository"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/service"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/compliance-service/pkg/crypto"
	"github.com/trigold786/92-Account-Center/compliance-service/pkg/mq"
	"github.com/trigold786/92-Account-Center/pkg/config"
	healthpkg "github.com/trigold786/92-Account-Center/pkg/health"
	"github.com/trigold786/92-Account-Center/pkg/logging"
)

var logger *slog.Logger

func init() {}

var (
	requestCount    uint64
	durationSumNano uint64
	durationCount   uint64
)

func main() {
	logger = logging.NewLogger("compliance-service")

	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "postgres")
	dbPassword := getEnvSecret("DB_PASSWORD", "postgres")
	dbName := getEnv("DB_NAME", "account_center")

	dsn := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser + " password=" + dbPassword + " dbname=" + dbName + " sslmode=disable"
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("failed to connect to database", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("connected to database")

	redisURL := getEnv("REDIS_URL", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: getEnvSecret("REDIS_PASSWORD", ""),
		DB:       0,
	})
	defer rdb.Close()

	var healthCheckers []healthpkg.Checker
	if db != nil {
		healthCheckers = append(healthCheckers, &healthpkg.PostgresChecker{
			Ping: func(ctx context.Context) error {
				_, err := db.ExecContext(ctx, "SELECT 1")
				return err
			},
		})
	}
	if rdb != nil {
		healthCheckers = append(healthCheckers, &healthpkg.RedisChecker{
			Ping: func(ctx context.Context) error {
				return rdb.Ping(ctx).Err()
			},
		})
	}
	compositeHealth := healthpkg.CompositeChecker{Checkers: healthCheckers}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	configURL := getEnv("CONFIG_SERVICE_URL", "http://localhost:30315")
	configClient := config.NewClient(configURL)
	svcCfg, err := svcconfig.Load(configClient)
	if err != nil {
	logger.Warn("config-service unavailable, continuing with env/defaults", "error", err)
}
	logger.Info("compliance config loaded successfully")

	riskRepo := repository.NewRiskRepository(db)
	geoService := service.NewGeoService()
	riskService := service.NewRiskService(riskRepo, geoService, svcCfg)
	riskHandler := handler.NewRiskHandler(riskService)

	auditRepo := repository.NewAuditRepository(db)
	auditService := service.NewAuditService(auditRepo, svcCfg)
	auditHandler := handler.NewAuditHandler(auditService)

	encryptKey, err := crypto.KeyFromEnv("ENCRYPTION_KEY")
	if err != nil {
	logger.Warn("config-service unavailable, continuing with env/defaults", "error", err)
}
	if os.Getenv("ENCRYPTION_KEY") == "" {
		logger.Warn("ENCRYPTION_KEY not set, using ephemeral key. KYB data will be lost on restart")
	}

	entRepo := repository.NewEnterpriseRepository(db)
	kybService := service.NewKYBService(entRepo, encryptKey)
	kybHandler := handler.NewKYBHandler(kybService)

	blacklistRepo := repository.NewBlacklistRepository(db)
	blacklistSvc := service.NewBlacklistService(blacklistRepo, rdb, svcCfg)
	blacklistHandler := handler.NewBlacklistHandler(blacklistSvc)
	windowLimiter := service.NewSlidingWindowLimiter(rdb, svcCfg)
	_ = windowLimiter

	r := gin.New()
	r.Use(gin.RecoveryWithWriter(os.Stderr, func(c *gin.Context, err any) {
		logger.Error("panic recovered", "error", fmt.Sprintf("%v", err))
	}))
	r.Use(logging.Middleware(logger))

	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		atomic.AddUint64(&requestCount, 1)
		atomic.AddUint64(&durationSumNano, uint64(time.Since(start).Nanoseconds()))
		atomic.AddUint64(&durationCount, 1)
	})

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

	blacklistGroup := r.Group("/api/v1/blacklist")
	{
		blacklistGroup.POST("/", blacklistHandler.AddEntry)
		blacklistGroup.POST("/check", blacklistHandler.CheckEntry)
		blacklistGroup.DELETE("/:type/:value", blacklistHandler.RemoveEntry)
		blacklistGroup.GET("/", blacklistHandler.ListEntries)
	}

	internalFraud := r.Group("/internal/v1/fraud")
	{
		internalFraud.POST("/check-registration", func(c *gin.Context) {
			var req struct {
				IP string `json:"ip" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "invalid request body"})
				return
			}
			blocked, reason, _ := blacklistSvc.CheckBlocked(c.Request.Context(), "IP", req.IP)
			if blocked {
				c.JSON(200, gin.H{"blocked": true, "reason": reason})
				return
			}
			allowed, count, _ := windowLimiter.CheckRegistrationLimit(c.Request.Context(), req.IP)
			c.JSON(200, gin.H{"blocked": !allowed, "current_count": count, "limit": svcCfg.SlidingWindowRegLimit})
		})
	}

	r.Any("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"compliance-service\"} %d\n", atomic.LoadUint64(&requestCount))
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_sum Total request duration in seconds\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_sum counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_sum{service=\"compliance-service\"} %f\n", time.Duration(atomic.LoadUint64(&durationSumNano)).Seconds())
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_count Total request count for duration\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_count counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_count{service=\"compliance-service\"} %d\n", atomic.LoadUint64(&durationCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"compliance-service\"} %d\n", runtime.NumGoroutine())
		c.Data(http.StatusOK, "text/plain; version=0.0.4", buf.Bytes())
	})

	r.Any("/health", func(c *gin.Context) {
		result := compositeHealth.Check(c.Request.Context())
		resp := healthpkg.BuildResponse(result.Checks)
		statusCode := 200
		if result.Status == healthpkg.StatusDown {
			statusCode = 503
		}
		c.JSON(statusCode, resp)
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
			logger.Warn("failed to create Kafka MQ, audit logs will be synchronous only", "error", err.Error())
		} else {
			messageQueue = kafkaMQ
		}

	case "redis":
		streamKey := getEnv("REDIS_STREAM_KEY", "audit-logs")
		groupName := getEnv("REDIS_CONSUMER_GROUP", "compliance-service")
		consumerID := getEnv("REDIS_CONSUMER_ID", "compliance-service-1")

		redisPassword := getEnvSecret("REDIS_PASSWORD", "")
		redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

		redisMQ := mq.NewRedisStreamsMQ(redisURL, redisPassword, redisDB, streamKey, groupName, consumerID)
		messageQueue = redisMQ

	default:
		logger.Warn("unknown AUDIT_MQ_TYPE, falling back to synchronous-only mode", "type", mqType)
	}

	if messageQueue != nil {
		defer messageQueue.Close()
		if err := messageQueue.StartConsumer(ctx, auditService); err != nil {
			logger.Warn("failed to start MQ consumer, audit logs will be synchronous only", "error", err.Error())
		} else {
			logger.Info("MQ consumer started", "type", mqType)
		}
	} else {
		logger.Info("no message queue configured, audit logs are synchronous only")
	}

	port := getEnv("PORT", "30313")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		defer logging.RecoverGoroutine(logger, "signal-handler")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("shutting down")
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Warn("server forced shutdown", "error", err.Error())
		}
	}()

	logger.Info("starting server", "port", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("failed to start server", "error", err.Error())
		os.Exit(1)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvSecret(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	logger.Warn("environment variable not set, using an insecure default", "key", key)
	return defaultValue
}
