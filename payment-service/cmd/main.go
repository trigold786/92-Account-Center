package main

import (
	"bytes"
	"context"
	"crypto/subtle"
	"database/sql"
	"encoding/json"
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
	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/provider"
	"github.com/trigold786/92-Account-Center/payment-service/internal/repository"
	"github.com/trigold786/92-Account-Center/payment-service/internal/service"
	"github.com/trigold786/92-Account-Center/payment-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/payment-service/internal/worker"
	"github.com/trigold786/92-Account-Center/pkg/config"
	"github.com/trigold786/92-Account-Center/pkg/env"
	healthpkg "github.com/trigold786/92-Account-Center/pkg/health"
	"github.com/trigold786/92-Account-Center/pkg/logging"
)

var logger = slog.Default()

type httpNotifier struct {
	baseURL string
	client  *http.Client
}

func newHTTPNotifier(baseURL string) *httpNotifier {
	return &httpNotifier{baseURL: baseURL, client: &http.Client{Timeout: 5 * time.Second}}
}

func (n *httpNotifier) CancelRefundedOrderSubscription(ctx context.Context, userID int64, orderID int64, reason string) error {
	payload, err := json.Marshal(struct {
		UserID  int64  `json:"user_id"`
		OrderID string `json:"order_id"`
		Reason  string `json:"reason"`
	}{UserID: userID, OrderID: fmt.Sprintf("%d", orderID), Reason: reason})
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, n.baseURL+"/internal/v1/subscriptions/cancel-refunded-order", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("account-service refund cancellation returned status %d", resp.StatusCode)
	}
	return nil
}

func (n *httpNotifier) NotifySubscriptionActivation(ctx context.Context, req handler.SubscriptionActivationRequest) error {
	payload, err := json.Marshal(req)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, n.baseURL+"/internal/v1/subscriptions/activate-paid-order", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("account-service activation returned status %d", resp.StatusCode)
	}
	return nil
}

type httpCreditService struct {
	baseURL string
	client  *http.Client
}

func newHTTPCreditService(baseURL string) *httpCreditService {
	return &httpCreditService{baseURL: baseURL, client: &http.Client{Timeout: 5 * time.Second}}
}

func (s *httpCreditService) ReverseCredits(ctx context.Context, userID int64, amount int, reason string) error {
	return nil
}

type orderRepoExpiryAdapter struct {
	repo repository.OrderRepository
}

func (a *orderRepoExpiryAdapter) FindExpired(ctx context.Context, before time.Time) ([]worker.Order, error) {
	orders, err := a.repo.FindExpired(ctx, before)
	if err != nil {
		return nil, err
	}
	result := make([]worker.Order, len(orders))
	for i, o := range orders {
		result[i] = worker.Order{ID: o.ID, Status: string(o.Status)}
	}
	return result, nil
}

func (a *orderRepoExpiryAdapter) UpdateStatus(ctx context.Context, id int64, status string) error {
	return a.repo.UpdateStatus(ctx, id, model.OrderStatus(status), "", "")
}

type orderRepoRefundAdapter struct {
	repo repository.OrderRepository
}

func (a *orderRepoRefundAdapter) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	return a.repo.GetByID(ctx, id)
}

func (a *orderRepoRefundAdapter) GetPendingOrdersOlderThan(ctx context.Context, since time.Duration) ([]*model.Order, error) {
	return a.repo.GetPendingOrdersOlderThan(ctx, since)
}

func (a *orderRepoRefundAdapter) UpdateStatus(ctx context.Context, id int64, status string) error {
	return a.repo.UpdateStatus(ctx, id, model.OrderStatus(status), "", "")
}

func (a *orderRepoRefundAdapter) UpdateOrderStatus(ctx context.Context, orderNo string, fromStatus, toStatus string) error {
	return a.repo.UpdateOrderStatus(ctx, orderNo, fromStatus, toStatus)
}

var (
	requestCount    uint64
	durationSumNano uint64
	durationCount   uint64
)

func main() {
	logger = logging.NewLogger("payment-service")

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

	configURL := env.Get("CONFIG_SERVICE_URL", "http://localhost:30315")
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
	invoiceSvc := service.NewInvoiceService(nil)
	invoiceSvcHandler := handler.NewInvoiceServiceHandler(invoiceSvc)
	paymentFlowHandler := handler.NewPaymentFlowHandler()

	providerRegistry := provider.NewProviderRegistry()
	paymentMode := env.Get("PAYMENT_MODE", "sandbox")
	wechatCfg := service.WeChatPayConfig{
		Mode:                paymentMode,
		AppID:               env.Get("WECHAT_APP_ID", "wx_sandbox_app_id"),
		MchID:               env.Get("WECHAT_MCH_ID", "sandbox_mch_id"),
		APIKey:              env.GetSecret("WECHAT_API_KEY", "sandbox_api_key"),
		CertificateSerialNo: env.Get("WECHAT_CERT_SERIAL_NO", ""),
		PrivateKeyPath:      env.Get("WECHAT_PRIVATE_KEY_PATH", ""),
	}
	if err := wechatCfg.ValidateProduction(); err != nil {
		logger.Error("invalid wechat production payment config", "error", err)
		os.Exit(1)
	}
	wechatProvider := service.NewWeChatPayProvider(wechatCfg)
	alipayCfg := service.AlipayConfig{
		Mode:       paymentMode,
		AppID:      env.Get("ALIPAY_APP_ID", "alipay_sandbox_app_id"),
		PrivateKey: env.GetSecret("ALIPAY_PRIVATE_KEY", "sandbox_private_key"),
		PublicKey:  env.GetSecret("ALIPAY_PUBLIC_KEY", "sandbox_public_key"),
		NotifyURL:  env.Get("ALIPAY_NOTIFY_URL", "http://localhost:30316/api/v1/payment/callback/alipay"),
	}
	if err := alipayCfg.ValidateProduction(); err != nil {
		logger.Error("invalid alipay production payment config", "error", err)
		os.Exit(1)
	}
	alipayProvider := service.NewAlipayProvider(alipayCfg)
	providerRegistry.Register(wechatProvider)
	providerRegistry.Register(alipayProvider)

	refundRepo := repository.NewRefundRepository(db)
	creditSvc := newHTTPCreditService(env.Get("CREDIT_SERVICE_URL", "http://localhost:30312"))
	accountNotifier := newHTTPNotifier(env.Get("ACCOUNT_SERVICE_URL", "http://localhost:30301"))
	orderRepoRefund := &orderRepoRefundAdapter{repo: orderRepo}
	refundSvc := service.NewRefundService(refundRepo, orderRepoRefund, creditSvc, accountNotifier, providerRegistry)
	refundHandler := handler.NewRefundHandler(refundSvc)

	paymentCallbackRepo := repository.NewPaymentCallbackRepository(db)
	callbackHandler := handler.NewCallbackHandlerWithActivationNotifier(providerRegistry, orderRepo, orderSvc, paymentCallbackRepo, accountNotifier, logger)
	createPaymentHandler := handler.NewCreatePaymentHandler(providerRegistry, orderSvc, logger)
	reconciliationReportRepo := service.NewSQLReconciliationReportRepository(db)
	reconciliationSvc := service.NewReconciliationServiceWithRepository(providerRegistry, orderRepo, reconciliationReportRepo, logger)

	expiryInterval := 5 * time.Minute
	if svcCfg != nil && svcCfg.OrderExpiryMinutes > 0 {
		expiryInterval = time.Duration(svcCfg.OrderExpiryMinutes) * time.Minute
	}
	orderExpiryRepo := &orderRepoExpiryAdapter{repo: orderRepo}
	expiryWorker := worker.NewOrderExpiryWorker(orderExpiryRepo, expiryInterval)
	expiryCtx, expiryCancel := context.WithCancel(context.Background())
	defer expiryCancel()
	go expiryWorker.Start(expiryCtx)

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
		invoiceGroup.GET("/:id", invoiceSvcHandler.GetInvoiceSvc)
		invoiceGroup.POST("/svc", invoiceSvcHandler.CreateInvoiceSvc)
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
				logger.Error("reconciliation failed", "error", err)
				c.JSON(500, gin.H{"code": 500, "message": "reconciliation failed"})
				return
			}
			c.JSON(200, gin.H{"code": 200, "message": "success", "data": report})
		})
		paymentGroup.GET("/reconcile/:report_id", func(c *gin.Context) {
			reportID := c.Param("report_id")
			report, err := reconciliationSvc.GetReconciliationReport(c.Request.Context(), reportID)
			if err != nil {
				logger.Error("get reconciliation report failed", "error", err)
				c.JSON(404, gin.H{"code": 404, "message": "report not found"})
				return
			}
			c.JSON(200, gin.H{"code": 200, "message": "success", "data": report})
		})
	}

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
		showDetails := env.Get("HEALTH_SHOW_DETAILS", "false") == "true"
		resp := healthpkg.BuildResponseConditional(result.Checks, showDetails)
		statusCode := 200
		if result.Status == healthpkg.StatusDown {
			statusCode = 503
		}
		c.JSON(statusCode, resp)
	})

	port := env.Get("PORT", "30316")
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


