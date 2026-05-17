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
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	_ "github.com/lib/pq"

	"github.com/trigold786/92-Account-Center/account-service/internal/cache"
	"github.com/trigold786/92-Account-Center/account-service/internal/handler"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
	"github.com/trigold786/92-Account-Center/account-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/account-service/pkg/sms"
	"github.com/trigold786/92-Account-Center/pkg/config"
	"github.com/trigold786/92-Account-Center/pkg/logging"
)

var (
	requestCount    uint64
	durationSumNano uint64
	durationCount   uint64
)

var logger = logging.NewLogger("account-service")


func main() {
	slog.SetDefault(logger)
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

	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("failed to parse REDIS_URL", "error", err.Error())
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	logger.Info("connected to redis")

	configURL := getEnv("CONFIG_SERVICE_URL", "http://localhost:30315")
	configClient := config.NewClient(configURL)
	svcCfg, err := svcconfig.Load(configClient)
	if err != nil {
	logger.Warn("config-service unavailable, continuing with env/defaults", "error", err)
}
	logger.Info("config loaded")

	userRepo := repository.NewUserRepository(db)
	entitlementRepo := repository.NewEntitlementRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	smsClient := sms.NewClient(getEnv("SMS_SERVICE_URL", "http://localhost:30311"))

	entitlementCache := cache.NewEntitlementCache(rdb, svcCfg.EntitlementCacheTTL)

	userService := service.NewUserService(userRepo, smsClient, svcCfg)
	entitlementService := service.NewEntitlementService(entitlementRepo, entitlementCache)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, userRepo, entitlementService, svcCfg)

	var referralBinder handler.ReferralBinder
	creditServiceURL := getEnv("CREDIT_SERVICE_URL", "")
	if creditServiceURL != "" {
		referralBinder = service.NewReferralClient(creditServiceURL)
	}

	registerHandler := handler.NewRegisterHandler(userService, referralBinder)
	passwordHandler := handler.NewPasswordHandler(userService)
	deletionHandler := handler.NewDeletionHandler(userService)
	tierHandler := handler.NewTierHandler(userRepo)
	entitlementHandler := handler.NewEntitlementHandler(entitlementService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)

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

	accountGroup := r.Group("/api/v1/account")
	{
		accountGroup.POST("/register", registerHandler.Register)

		accountGroup.POST("/password/send-verification-code", passwordHandler.SendVerificationCode)
		accountGroup.POST("/password/change", passwordHandler.ChangePassword)

		accountGroup.POST("/deletion/request", deletionHandler.RequestDeletion)
		accountGroup.POST("/deletion/cancel", deletionHandler.CancelDeletion)
		accountGroup.GET("/deletion/status", deletionHandler.GetDeletionStatus)

		accountGroup.GET("/:user_id/tier", tierHandler.GetTier)
	}

	internalAccountGroup := r.Group("/internal/v1/account")
	{
		internalAccountGroup.PUT("/:user_id/tier", tierHandler.UpdateTier)
	}

	entitlementGroup := r.Group("/api/v1/entitlements")
	{
		entitlementGroup.GET("/:user_id", entitlementHandler.GetUserEntitlements)
	}

	internalEntitlementGroup := r.Group("/internal/v1/entitlements")
	{
		internalEntitlementGroup.POST("/consume", entitlementHandler.Consume)
		internalEntitlementGroup.POST("/grant", entitlementHandler.Grant)
	}

	subscriptionGroup := r.Group("/api/v1/subscriptions")
	{
		subscriptionGroup.POST("/purchase", subscriptionHandler.Purchase)
		subscriptionGroup.POST("/upgrade", subscriptionHandler.Upgrade)
		subscriptionGroup.POST("/renew", subscriptionHandler.Renew)
		subscriptionGroup.GET("/:user_id", subscriptionHandler.GetUserSubscriptions)
	}

	r.Any("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"account-service\"} %d\n", atomic.LoadUint64(&requestCount))
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_sum Total request duration in seconds\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_sum counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_sum{service=\"account-service\"} %f\n", time.Duration(atomic.LoadUint64(&durationSumNano)).Seconds())
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_count Total request count for duration\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_count counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_count{service=\"account-service\"} %d\n", atomic.LoadUint64(&durationCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"account-service\"} %d\n", runtime.NumGoroutine())
		c.Data(http.StatusOK, "text/plain; version=0.0.4", buf.Bytes())
	})

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30301")
	logger.Info("starting server", "port", port)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		defer logging.RecoverGoroutine(logger, "shutdown-listener")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("shutting down")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

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
	logger.Warn("environment variable not set, using insecure default", "key", key)
	return defaultValue
}