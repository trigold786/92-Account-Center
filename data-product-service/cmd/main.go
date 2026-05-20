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

	"github.com/trigold786/92-Account-Center/data-product-service/internal/handler"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/repository"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/pkg/config"
	healthpkg "github.com/trigold786/92-Account-Center/pkg/health"
	"github.com/trigold786/92-Account-Center/pkg/logging"
)

var (
	requestCount    uint64
	durationSumNano uint64
	durationCount   uint64
)

var logger *slog.Logger

func init() {}

func main() {
	logger = logging.NewLogger("data-product-service")
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

	var replicaDB *sql.DB
	readReplicaURL := getEnv("READ_REPLICA_URL", "")
	if readReplicaURL != "" {
		replicaDB, err = sql.Open("postgres", readReplicaURL)
		if err != nil {
			logger.Warn("failed to connect to read replica, falling back to primary", "error", err.Error())
			replicaDB = nil
		} else if err := replicaDB.Ping(); err != nil {
			logger.Warn("read replica ping failed, falling back to primary", "error", err.Error())
			replicaDB = nil
		} else {
			logger.Info("connected to read replica")
		}
	}

	readReplica := repository.NewReadReplicaRepo(db, replicaDB)
	_ = readReplica

	var healthCheckers []healthpkg.Checker
	if db != nil {
		healthCheckers = append(healthCheckers, &healthpkg.PostgresChecker{
			Ping: func(ctx context.Context) error {
				_, err := db.ExecContext(ctx, "SELECT 1")
				return err
			},
		})
	}
	compositeHealth := healthpkg.CompositeChecker{Checkers: healthCheckers}

	configURL := getEnv("CONFIG_SERVICE_URL", "http://localhost:30315")
	configClient := config.NewClient(configURL)
	svcCfg, err := svcconfig.Load(configClient)
	if err != nil {
	logger.Warn("config-service unavailable, continuing with env/defaults", "error", err)
}
	logger.Info("data-product config loaded successfully")

	dataRepo := repository.NewDataRepository(db)
	rfmSvc := service.NewRFMService(dataRepo)
	dashSvc := service.NewDashboardService(dataRepo, rfmSvc, svcCfg)

	rfmHandler := handler.NewRFMHandler(rfmSvc)
	dashHandler := handler.NewDashboardHandler(dashSvc)
	funnelHandler := handler.NewFunnelHandler(dashSvc)

	eventRepo := repository.NewEventRepository(db)
	eventSvc := service.NewEventService(eventRepo)
	eventHandler := handler.NewEventHandler(eventSvc)

	metricsRepo := repository.NewMetricsRepository(db)
	metricsSvc := service.NewMetricsService(metricsRepo)
	opsDashHandler := handler.NewOpsDashboardHandler(metricsSvc)

	streamSvc := service.NewStreamService(nil)
	streamHandler := handler.NewStreamHandler(streamSvc)

	adEventSvc := service.NewAdEventService(nil)
	adEventHandler := handler.NewAdEventHandler(adEventSvc)

	abTestSvc := service.NewABTestService()
	abTestHandler := handler.NewABTestHandler(abTestSvc)

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

	dataGroup := r.Group("/api/v1/data")
	{
		dataGroup.GET("/rfm/:user_id", rfmHandler.GetRFM)
		dataGroup.POST("/rfm/batch", rfmHandler.GetRFMBatch)
		dataGroup.GET("/dashboard/overview", dashHandler.GetOverview)
		dataGroup.GET("/funnel/subscription", funnelHandler.GetSubscriptionFunnel)
	}

	eventsGroup := r.Group("/api/v1/events")
	{
		eventsGroup.POST("", eventHandler.TrackEvent)
		eventsGroup.POST("/batch", eventHandler.BatchTrack)
	}

	opsGroup := r.Group("/api/v1/ops")
	{
		opsGroup.GET("/registration-trends", opsDashHandler.GetRegistrationTrends)
		opsGroup.GET("/conversion-funnel", opsDashHandler.GetConversionFunnel)
		opsGroup.GET("/mrr", opsDashHandler.GetMRR)
		opsGroup.GET("/k-factor", opsDashHandler.GetKFactor)
		opsGroup.GET("/rfm-distribution", opsDashHandler.GetRFM)
	}

	streamGroup := r.Group("/api/v1/stream")
	{
		streamGroup.POST("/events", streamHandler.ProcessEvent)
		streamGroup.GET("/online", streamHandler.GetOnlineCount)
		streamGroup.GET("/funnel", streamHandler.GetRealtimeFunnel)
	}

	adGroup := r.Group("/api/v1/ad")
	{
		adGroup.POST("/events", adEventHandler.TrackAdEvent)
		adGroup.GET("/metrics", adEventHandler.GetAdMetrics)
	}

	experimentGroup := r.Group("/api/v1/experiments")
	{
		experimentGroup.POST("", abTestHandler.CreateExperiment)
		experimentGroup.GET("/:id/assign", abTestHandler.AssignVariant)
		experimentGroup.POST("/:id/events", abTestHandler.RecordEvent)
		experimentGroup.GET("/:id/results", abTestHandler.GetResults)
	}

	r.Any("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"data-product-service\"} %d\n", atomic.LoadUint64(&requestCount))
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_sum Total request duration in seconds\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_sum counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_sum{service=\"data-product-service\"} %f\n", time.Duration(atomic.LoadUint64(&durationSumNano)).Seconds())
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_count Total request count for duration\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_count counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_count{service=\"data-product-service\"} %d\n", atomic.LoadUint64(&durationCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"data-product-service\"} %d\n", runtime.NumGoroutine())
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

	port := getEnv("PORT", "30314")
	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		defer logging.RecoverGoroutine(logger, "shutdown")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("shutting down server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
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
	logger.Warn("environment variable not set, using insecure default", "key", key)
	return defaultValue
}
