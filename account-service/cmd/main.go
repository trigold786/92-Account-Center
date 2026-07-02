package main

import (
	"bytes"
	"context"
	"crypto/subtle"
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
	"github.com/hibiken/asynq"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/account-service/internal/cache"
	"github.com/trigold786/92-Account-Center/account-service/internal/handler"
	"github.com/trigold786/92-Account-Center/account-service/internal/middleware"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
	"github.com/trigold786/92-Account-Center/account-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/account-service/internal/worker"
	"github.com/trigold786/92-Account-Center/account-service/pkg/sms"
	"github.com/trigold786/92-Account-Center/pkg/config"
	"github.com/trigold786/92-Account-Center/pkg/env"
	healthpkg "github.com/trigold786/92-Account-Center/pkg/health"
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
	dbHost := env.Get("DB_HOST", "localhost")
	dbPort := env.Get("DB_PORT", "5432")
	dbUser := env.Get("DB_USER", "postgres")
	dbPassword := env.GetSecret("DB_PASSWORD", "postgres")
	dbName := env.Get("DB_NAME", "account_center")
	sslmode := env.Get("DB_SSLMODE", "disable")

	dsn := "host=" + dbHost + " port=" + dbPort + " user=" + dbUser + " password=" + dbPassword + " dbname=" + dbName + " sslmode=" + sslmode
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		logger.Error("failed to connect to database", "error", err.Error())
		os.Exit(1)
	}
	defer db.Close()

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		logger.Error("failed to ping database", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("connected to database")

	redisURL := env.Get("REDIS_URL", "redis://localhost:6379")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("failed to parse REDIS_URL", "error", err.Error())
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	logger.Info("connected to redis")

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

	configURL := env.Get("CONFIG_SERVICE_URL", "http://localhost:30315")
	configClient := config.NewClient(configURL)
	svcCfg, err := svcconfig.Load(configClient)
	if err != nil {
		logger.Warn("config-service unavailable, continuing with env/defaults", "error", err)
	}
	logger.Info("config loaded")

	userRepo := repository.NewUserRepository(db)
	entitlementRepo := repository.NewEntitlementRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	smsClient := sms.NewClient(env.Get("SMS_SERVICE_URL", "http://localhost:30311"))

	entitlementCache := cache.NewEntitlementCache(rdb, svcCfg.EntitlementCacheTTL)

	userService := service.NewUserService(userRepo, smsClient, svcCfg)
	entitlementService := service.NewEntitlementService(entitlementRepo, entitlementCache)
	paymentVerifier := service.NewHTTPPaymentOrderVerifier(env.Get("PAYMENT_SERVICE_URL", "http://localhost:30316"))
	subscriptionService := service.NewSubscriptionServiceWithPaymentVerifier(subscriptionRepo, userRepo, entitlementService, svcCfg, paymentVerifier)

	sessionInvalidator := service.NewHTTPSessionInvalidator(env.Get("AUTH_SERVICE_URL", "http://localhost:30302"))
	deletionService := service.NewDeletionService(userRepo, entitlementRepo, rdb, logger, sessionInvalidator)
	deletionWorker := worker.NewDeletionWorker(deletionService, logger)

	renewalSvc := service.NewRenewalService(nil, nil)
	renewalWorker := worker.NewRenewalWorker(renewalSvc)

	redisAddr := env.Get("REDIS_ADDR", "localhost:6379")
	asynqServer := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 5},
	)
	asynqMux := worker.NewServeMux(deletionWorker, renewalWorker)

	scheduler, err := worker.NewScheduler(redisAddr)
	if err != nil {
		logger.Error("failed to create scheduler", "error", err.Error())
		os.Exit(1)
	}

	go func() {
		defer logging.RecoverGoroutine(logger, "asynq-worker")
		logger.Info("starting asynq worker")
		if err := asynqServer.Run(asynqMux); err != nil {
			logger.Error("asynq worker stopped", "error", err.Error())
		}
	}()

	go func() {
		defer logging.RecoverGoroutine(logger, "asynq-scheduler")
		logger.Info("starting asynq scheduler")
		if err := scheduler.Run(); err != nil {
			logger.Error("asynq scheduler stopped", "error", err.Error())
		}
	}()

	var referralBinder handler.ReferralBinder
	creditServiceURL := env.Get("CREDIT_SERVICE_URL", "")
	if creditServiceURL != "" {
		referralBinder = service.NewReferralClient(creditServiceURL)
	}

	registerHandler := handler.NewRegisterHandler(userService, referralBinder)
	passwordHandler := handler.NewPasswordHandler(userService)
	deletionHandler := handler.NewDeletionHandler(userService)
	tierHandler := handler.NewTierHandler(userRepo)
	entitlementHandler := handler.NewEntitlementHandler(entitlementService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)
	pricingHandler := handler.NewPricingHandler()

	dashboardSvc := service.NewDashboardService(nil)
	dashboardHandler := handler.NewDashboardHandler(dashboardSvc)

	upgradeSvc := service.NewUpgradeService(nil, nil)
	upgradeHandler := handler.NewUpgradeHandler(upgradeSvc)

	subAdminRepo := repository.NewSubscriptionAdminRepository(db)
	subAdminSvc := service.NewSubscriptionAdminService(subAdminRepo)
	subAdminHandler := handler.NewSubscriptionAdminHandler(subAdminSvc)

	exportSvc := service.NewExportService(nil, "")
	exportHandler := handler.NewExportHandler(exportSvc)

	openAPISvc := service.NewOpenAPIService()
	openAPIHandler := handler.NewOpenAPIHandler(openAPISvc)

	searchSvc := service.NewSearchService(nil)
	searchHandler := handler.NewSearchHandler(searchSvc)

	leaderboardSvc := service.NewLeaderboardService(nil)
	leaderboardHandler := handler.NewLeaderboardHandler(leaderboardSvc)

	adminRepo := repository.NewAdminRepository(db)
	var creditClient service.CreditClient
	if creditServiceURL != "" {
		creditClient = service.NewHTTPCreditClient(creditServiceURL)
	}
	adminService := service.NewAdminService(adminRepo, creditClient)
	adminHandler := handler.NewAdminHandler(adminService)

	if err := adminRepo.EnsureAuditLogTable(context.Background()); err != nil {
		logger.Warn("failed to ensure audit_log table", "error", err)
	}

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

	internalAuth := func(c *gin.Context) {
		token := c.GetHeader("X-Internal-Token")
		expected := env.Get("INTERNAL_API_TOKEN", "")
		if expected == "" {
			c.JSON(http.StatusForbidden, gin.H{"error": "internal API not configured"})
			c.Abort()
			return
		}
		if subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}

	internalAccountGroup := r.Group("/internal/v1/account", internalAuth)
	{
		internalAccountGroup.PUT("/:user_id/tier", tierHandler.UpdateTier)
	}

	entitlementGroup := r.Group("/api/v1/entitlements")
	{
		entitlementGroup.GET("/:user_id", entitlementHandler.GetUserEntitlements)
	}

	internalEntitlementGroup := r.Group("/internal/v1/entitlements", internalAuth)
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

	internalSubscriptionGroup := r.Group("/internal/v1/subscriptions", internalAuth)
	{
		internalSubscriptionGroup.POST("/activate-paid-order", subscriptionHandler.ActivatePaidOrder)
		internalSubscriptionGroup.POST("/cancel-refunded-order", subscriptionHandler.CancelRefundedOrder)
	}

	pricingGroup := r.Group("/api/v1/pricing")
	{
		pricingGroup.GET("", pricingHandler.GetPricing)
		pricingGroup.POST("/calculate-discount", pricingHandler.CalculateDiscount)
	}

	r.GET("/api/v1/dashboard", dashboardHandler.GetDashboard)

	upgradeGroup := r.Group("/api/v1/upgrade")
	{
		upgradeGroup.POST("/preview", upgradeHandler.PreviewUpgrade)
		upgradeGroup.POST("/downgrade/preview", upgradeHandler.PreviewDowngrade)
		upgradeGroup.POST("/execute", upgradeHandler.ExecuteUpgrade)
	}

	jwtSecret := env.Get("JWT_ACCESS_SECRET", "default-secret")

	adminGroup := r.Group("/api/v1/admin")
	adminGroup.Use(middleware.AdminAuthMiddleware(jwtSecret))
	{
		adminGroup.GET("/users", adminHandler.ListUsers)
		adminGroup.GET("/users/:user_id", adminHandler.GetUserDetail)
		adminGroup.PUT("/users/:user_id/status", adminHandler.UpdateUserStatus)
		adminGroup.PUT("/users/:user_id/tier", adminHandler.AdjustIdentityTier)
		adminGroup.POST("/users/:user_id/credits", adminHandler.AdjustCredits)
		adminGroup.GET("/users/:user_id/audit-log", adminHandler.GetAuditLog)

		adminGroup.GET("/kyc/pending", adminHandler.ListPendingKYC)
		adminGroup.PUT("/kyc/:enterprise_id/review", adminHandler.ReviewKYC)

		adminGroup.POST("/plans", subAdminHandler.CreatePlan)
		adminGroup.GET("/plans", subAdminHandler.ListPlans)
		adminGroup.DELETE("/plans/:id", subAdminHandler.DeletePlan)

		adminGroup.POST("/coupons", subAdminHandler.CreateCoupon)
		adminGroup.GET("/coupons", subAdminHandler.ListCoupons)
	}

	exportGroup := r.Group("/api/v1/export")
	{
		exportGroup.GET("/personal", exportHandler.ExportPersonalData)
		exportGroup.POST("/request", exportHandler.RequestExport)
	}

	openAPIGroup := r.Group("/api/v1/openapi")
	{
		openAPIGroup.POST("/token", openAPIHandler.IssueToken)
	}

	r.GET("/api/v1/search", searchHandler.Search)
	r.GET("/api/v1/quick-actions", searchHandler.QuickActions)

	r.GET("/api/v1/leaderboard", leaderboardHandler.GetLeaderboard)
	r.GET("/api/v1/leaderboard/me", leaderboardHandler.GetMyRank)

	metricsAuth := func(c *gin.Context) {
		token := c.GetHeader("X-Internal-Token")
		expected := env.Get("INTERNAL_API_TOKEN", "")
		if expected == "" || subtle.ConstantTimeCompare([]byte(token), []byte(expected)) != 1 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
	r.Any("/metrics", metricsAuth, func(c *gin.Context) {
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
		result := compositeHealth.Check(c.Request.Context())
		showDetails := env.Get("HEALTH_SHOW_DETAILS", "false") == "true"
		resp := healthpkg.BuildResponseConditional(result.Checks, showDetails)
		statusCode := 200
		if result.Status == healthpkg.StatusDown {
			statusCode = 503
		}
		c.JSON(statusCode, resp)
	})

	port := env.Get("PORT", "30301")
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
		scheduler.Shutdown()
		asynqServer.Shutdown()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("failed to start server", "error", err.Error())
		os.Exit(1)
	}
}


