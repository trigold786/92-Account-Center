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

	"github.com/trigold786/92-Account-Center/payment-service/internal/handler"
	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
	"github.com/trigold786/92-Account-Center/payment-service/internal/repository"
	"github.com/trigold786/92-Account-Center/payment-service/internal/service"
	"github.com/trigold786/92-Account-Center/payment-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/pkg/config"
	healthpkg "github.com/trigold786/92-Account-Center/pkg/health"
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
	logger = logging.NewLogger("payment-service")

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
	logger.Info("payment config loaded successfully")

	orderRepo := repository.NewOrderRepository(db)
	orderSvc := service.NewOrderService(orderRepo, svcCfg)
	orderHandler := handler.NewOrderHandler(orderSvc)

	invoiceRepo := repository.NewInvoiceRepository(db)
	invoiceHandler := handler.NewInvoiceHandler(invoiceRepo)
	paymentFlowHandler := handler.NewPaymentFlowHandler()

	refundRepo := repository.NewRefundRepository(db)
	refundSvc := service.NewRefundService(refundRepo, nil, nil)
	refundHandler := handler.NewRefundHandler(refundSvc)

	providerRegistry := provider.NewProviderRegistry()
	wechatProvider := service.NewWeChatPayProvider(service.WeChatPayConfig{
		AppID:    getEnv("WECHAT_APP_ID", "wx_sandbox_app_id"),
		MchID:    getEnv("WECHAT_MCH_ID", "sandbox_mch_id"),
		APIKey:   getEnvSecret("WECHAT_API_KEY", "sandbox_api_key"),
		CertPath: getEnv("WECHAT_CERT_PATH", ""),
	})
	alipayProvider := service.NewAlipayProvider(service.AlipayConfig{
		AppID:      getEnv("ALIPAY_APP_ID", "alipay_sandbox_app_id"),
		PrivateKey: getEnvSecret("ALIPAY_PRIVATE_KEY", "sandbox_private_key"),
		PublicKey:  getEnvSecret("ALIPAY_PUBLIC_KEY", "sandbox_public_key"),
		NotifyURL:  getEnv("ALIPAY_NOTIFY_URL", "http://localhost:30316/api/v1/payment/callback/alipay"),
	})
	providerRegistry.Register(wechatProvider)
	providerRegistry.Register(alipayProvider)

	callbackHandler := handler.NewCallbackHandler(providerRegistry, orderRepo, orderSvc, logger)
	createPaymentHandler := handler.NewCreatePaymentHandler(providerRegistry, orderSvc, logger)
	reconciliationSvc := service.NewReconciliationService(providerRegistry, orderRepo, logger)

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

	ordersGroup := r.Group("/api/v1/orders")
	{
		ordersGroup.POST("", orderHandler.CreateOrder)
		ordersGroup.GET("/:id", orderHandler.GetOrder)
		ordersGroup.GET("", orderHandler.ListOrders)
		ordersGroup.PUT("/:id/status", orderHandler.UpdateStatus)
		ordersGroup.GET("/export/csv", orderHandler.ExportCSV)
	}

	invoiceGroup := r.Group("/api/v1/invoices")
	{
		invoiceGroup.POST("", invoiceHandler.CreateInvoice)
		invoiceGroup.GET("", invoiceHandler.ListInvoices)
	}

	paymentFlowGroup := r.Group("/api/v1/payment-flow")
	{
		paymentFlowGroup.GET("/result/:order_no", paymentFlowHandler.GetPaymentResult)
		paymentFlowGroup.POST("/retry/:order_no", paymentFlowHandler.RetryPayment)
	}

	refundGroup := r.Group("/api/v1/refunds")
	{
		refundGroup.POST("", refundHandler.RequestRefund)
		refundGroup.PUT("/:id/approve", refundHandler.ApproveRefund)
		refundGroup.PUT("/:id/reject", refundHandler.RejectRefund)
	}

	paymentGroup := r.Group("/api/v1/payment")
	{
		paymentGroup.POST("/callback/:provider", callbackHandler.HandleCallback)
		paymentGroup.POST("/create", createPaymentHandler.CreatePayment)
		paymentGroup.POST("/reconcile", func(c *gin.Context) {
			var req struct {
				Provider string `json:"provider" binding:"required"`
				Date     string `json:"date" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"code": 400, "message": "invalid request body"})
				return
			}
			report, err := reconciliationSvc.ReconcileOrders(c.Request.Context(), req.Provider, req.Date)
			if err != nil {
				c.JSON(500, gin.H{"code": 500, "message": err.Error()})
				return
			}
			c.JSON(200, gin.H{"code": 200, "message": "success", "data": report})
		})
		paymentGroup.GET("/reconcile/:report_id", func(c *gin.Context) {
			reportID := c.Param("report_id")
			report, err := reconciliationSvc.GetReconciliationReport(c.Request.Context(), reportID)
			if err != nil {
				c.JSON(404, gin.H{"code": 404, "message": err.Error()})
				return
			}
			c.JSON(200, gin.H{"code": 200, "message": "success", "data": report})
		})
	}

	r.Any("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"payment-service\"} %d\n", atomic.LoadUint64(&requestCount))
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_sum Total request duration in seconds\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_sum counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_sum{service=\"payment-service\"} %f\n", time.Duration(atomic.LoadUint64(&durationSumNano)).Seconds())
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_count Total request count for duration\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_count counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_count{service=\"payment-service\"} %d\n", atomic.LoadUint64(&durationCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"payment-service\"} %d\n", runtime.NumGoroutine())
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

	port := getEnv("PORT", "30316")
	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		defer logging.RecoverGoroutine(logger, "signal-handler")
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		logger.Info("shutting down")
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
