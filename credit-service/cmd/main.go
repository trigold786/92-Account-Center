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
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/credit-service/internal/handler"
	"github.com/trigold786/92-Account-Center/credit-service/internal/repository"
	"github.com/trigold786/92-Account-Center/credit-service/internal/service"
	"github.com/trigold786/92-Account-Center/credit-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/credit-service/internal/worker"
	"github.com/trigold786/92-Account-Center/pkg/config"
	"github.com/trigold786/92-Account-Center/pkg/logging"
)

var logger *slog.Logger

func init() { slog.SetDefault(logger) }

var (
	requestCount    uint64
	durationSumNano uint64
	durationCount   uint64
)

func main() {
	logger = logging.NewLogger("credit-service")

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

	redisAddr := getEnv("REDIS_URL", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		logger.Warn("redis not available", "error", err.Error())
	}

	configURL := getEnv("CONFIG_SERVICE_URL", "http://localhost:30315")
	configClient := config.NewClient(configURL)
	svcCfg, err := svcconfig.Load(configClient)
	if err != nil {
		logger.Error("failed to load credit config", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("credit config loaded successfully")

	creditRepo := repository.NewCreditRepository(db)
	referralRepo := repository.NewReferralRepository(db)

	creditSvc := service.NewCreditService(creditRepo, db, svcCfg)
	referralSvc := service.NewReferralService(referralRepo, svcCfg)
	rebateSvc := service.NewRebateService(creditRepo, referralRepo, creditSvc, svcCfg)

	subWorker := worker.NewSubscriptionWorker(rdb, rebateSvc, svcCfg)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go subWorker.Start(ctx)

	creditHandler := handler.NewCreditHandler(creditSvc)
	referralHandler := handler.NewReferralHandler(referralSvc)

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

	creditsGroup := r.Group("/api/v1/credits")
	{
		creditsGroup.GET("/:user_id/account", creditHandler.GetAccount)
		creditsGroup.GET("/:user_id/transactions", creditHandler.GetTransactions)
		creditsGroup.POST("/calculate-discount", creditHandler.CalculateDiscount)
	}

	internalCredits := r.Group("/internal/v1/credits")
	{
		internalCredits.POST("/earn", creditHandler.EarnCredits)
		internalCredits.POST("/consume", creditHandler.ConsumeCredits)
		internalCredits.POST("/refund", creditHandler.RefundCredits)
	}

	referralGroup := r.Group("/api/v1/referral")
	{
		referralGroup.POST("/bind", referralHandler.BindReferral)
		referralGroup.POST("/generate-link", referralHandler.GenerateLink)
		referralGroup.GET("/:user_id/summary", referralHandler.GetSummary)
	}

	r.Any("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"credit-service\"} %d\n", atomic.LoadUint64(&requestCount))
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_sum Total request duration in seconds\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_sum counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_sum{service=\"credit-service\"} %f\n", time.Duration(atomic.LoadUint64(&durationSumNano)).Seconds())
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_count Total request count for duration\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_count counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_count{service=\"credit-service\"} %d\n", atomic.LoadUint64(&durationCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"credit-service\"} %d\n", runtime.NumGoroutine())
		c.Data(http.StatusOK, "text/plain; version=0.0.4", buf.Bytes())
	})

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30312")
	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		defer logging.RecoverGoroutine(logger, "signal-handler")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("shutting down")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
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
