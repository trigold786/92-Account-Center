package main

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/pkg/config"
	"github.com/trigold786/92-Account-Center/pkg/logging"
	"github.com/trigold786/92-Account-Center/notification-service/internal/handler"
	"github.com/trigold786/92-Account-Center/notification-service/internal/provider"
	"github.com/trigold786/92-Account-Center/notification-service/internal/service"
	"github.com/trigold786/92-Account-Center/notification-service/internal/svcconfig"
	circuitbreaker "github.com/trigold786/92-Account-Center/pkg/circuitbreaker"
	healthpkg "github.com/trigold786/92-Account-Center/pkg/health"
)

var (
	requestCount    uint64
	durationSumNano uint64
	durationCount   uint64
)

var logger = slog.Default()

func main() {
	logger = logging.NewLogger("notification-service")
	redisAddr := getEnv("REDIS_URL", "localhost:6379")
	redisPassword := getEnvSecret("REDIS_PASSWORD", "")
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		logger.Error("failed to connect to Redis", "error", err.Error())
		os.Exit(1)
	}
	logger.Info("connected to Redis")

	var healthCheckers []healthpkg.Checker
	if redisClient != nil {
		healthCheckers = append(healthCheckers, &healthpkg.RedisChecker{
			Ping: func(ctx context.Context) error {
				return redisClient.Ping(ctx).Err()
			},
		})
	}
	compositeHealth := healthpkg.CompositeChecker{Checkers: healthCheckers}

	service.SetRedisClient(redisClient)

	configURL := getEnv("CONFIG_SERVICE_URL", "http://localhost:30315")
	configClient := config.NewClient(configURL)
	svcCfg, err := svcconfig.Load(configClient)
	if err != nil {
	logger.Warn("config-service unavailable, continuing with env/defaults", "error", err)
}
	logger.Info("notification config loaded successfully")

	aliyunProvider := provider.NewAliyunProvider(
		getEnv("ALIYUN_ACCESS_KEY_ID", ""),
		getEnvSecret("ALIYUN_ACCESS_KEY_SECRET", ""),
		getEnv("ALIYUN_SIGN_NAME", "速通互联验证码"),
	)
	tencentProvider := provider.NewTencentProvider(
		getEnv("TENCENT_APP_ID", ""),
		getEnvSecret("TENCENT_APP_SECRET", ""),
		getEnv("TENCENT_SIGN_NAME", "AccountCenter"),
	)
	chinaTelecomProvider := provider.NewChinaTelecomProvider(
		getEnv("CHINATELECOM_APP_ID", ""),
		getEnvSecret("CHINATELECOM_APP_SECRET", ""),
		getEnv("CHINATELECOM_SIGN_NAME", "AccountCenter"),
	)

	smsProviders := []provider.SMSProvider{aliyunProvider, tencentProvider, chinaTelecomProvider}
	circuitBreakers := []*circuitbreaker.CircuitBreaker{
		circuitbreaker.New(5, 30*time.Second),
		circuitbreaker.New(5, 30*time.Second),
		circuitbreaker.New(5, 30*time.Second),
	}

	smsService := service.NewSMSService(smsProviders, circuitBreakers, svcCfg)
	smsHandler := handler.NewSMSHandler(smsService)

	simpleSMTP := provider.NewSimpleSMTPProvider(
		getEnv("SMTP_HOST", "smtp.163.com"),
		getEnv("SMTP_PORT", "465"),
		getEnv("SMTP_USERNAME", ""),
		getEnvSecret("SMTP_PASSWORD", ""),
		getEnv("SMTP_FROM", ""),
	)
	verificationEmailService := service.NewSimpleEmailService(simpleSMTP, svcCfg)
	verificationEmailHandler := handler.NewVerificationEmailHandler(verificationEmailService)

	emailProviderType := getEnv("EMAIL_PROVIDER", "smtp")
	jwtSecret := getEnvSecret("JWT_SECRET", "your-secret-key-change-in-production")
	fromAddress := getEnv("FROM_ADDRESS", "noreply@accountcenter.com")

	var emailProvider provider.EmailProvider

	switch emailProviderType {
	case "sendgrid":
		emailProvider = provider.NewSendGridProvider(getEnv("SENDGRID_API_KEY", ""), fromAddress)
	case "aws_ses":
		sesProvider, sesErr := provider.NewSESProvider(
			getEnv("AWS_REGION", "us-east-1"),
			getEnv("AWS_ACCESS_KEY", ""),
			getEnvSecret("AWS_SECRET_KEY", ""),
			fromAddress,
		)
		if sesErr != nil {
			logger.Warn("failed to initialize SES provider", "error", sesErr)
		} else {
			emailProvider = sesProvider
		}
	case "smtp":
		emailProvider = provider.NewSMTPProvider(
			getEnv("SMTP_HOST", "localhost"),
			getEnv("SMTP_PORT", "587"),
			getEnv("SMTP_USERNAME", ""),
			getEnvSecret("SMTP_PASSWORD", ""),
			fromAddress,
		)
	default:
		logger.Error("unsupported email provider", "provider", emailProviderType)
		os.Exit(1)
	}
	logger.Info("using email provider", "provider", emailProvider.Name())

	otpEmailService := service.NewOTPEmailService(redisClient, emailProvider, jwtSecret, fromAddress, svcCfg)
	otpEmailHandler := handler.NewOTPEmailHandler(otpEmailService)

	pushRegistry := provider.NewPushProviderRegistry()
	pushRegistry.Register(provider.NewAPNsProvider(provider.APNsConfig{
		CertificatePath: getEnv("APNS_CERTIFICATE_PATH", ""),
		KeyPath:         getEnv("APNS_KEY_PATH", ""),
		BundleID:        getEnv("APNS_BUNDLE_ID", ""),
		Production:      getEnv("APNS_PRODUCTION", "false") == "true",
	}))
	pushRegistry.Register(provider.NewFCMProvider(provider.FCMConfig{
		ServerKey: getEnv("FCM_SERVER_KEY", ""),
		ProjectID: getEnv("FCM_PROJECT_ID", ""),
	}))
	pushRegistry.Register(provider.NewHMSProvider(provider.HMSConfig{
		AppID:     getEnv("HMS_APP_ID", ""),
		AppSecret: getEnvSecret("HMS_APP_SECRET", ""),
	}))
	logger.Info("push providers registered", "providers", pushRegistry.List())

	pushService := service.NewPushService(redisClient, pushRegistry)
	pushHandler := handler.NewPushHandler(pushService)
	deviceHandler := handler.NewDeviceHandler(pushService)

	wechatTemplateSvc := service.NewWeChatTemplateService(nil)
	wechatTemplateHandler := handler.NewWeChatTemplateHandler(wechatTemplateSvc)

	messageSvc := service.NewMessageService(nil)
	messageHandler := handler.NewMessageHandler(messageSvc)

	templateSvc := service.NewTemplateService(nil)
	templateHandler := handler.NewTemplateHandler(templateSvc)

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

	r.Any("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"notification-service\"} %d\n", atomic.LoadUint64(&requestCount))
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_sum Total request duration in seconds\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_sum counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_sum{service=\"notification-service\"} %f\n", time.Duration(atomic.LoadUint64(&durationSumNano)).Seconds())
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_count Total request count for duration\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_count counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_count{service=\"notification-service\"} %d\n", atomic.LoadUint64(&durationCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"notification-service\"} %d\n", runtime.NumGoroutine())
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

	smsGroup := r.Group("/api/v1/sms")
	{
		smsGroup.POST("/send", smsHandler.SendSMS)
		smsGroup.POST("/verify", smsHandler.VerifyCode)
		smsGroup.GET("/providers/status", smsHandler.GetProviderStatus)
	}

	emailGroup := r.Group("/api/v1/email")
	{
		emailGroup.POST("/verify", verificationEmailHandler.VerifyCode)
		emailGroup.POST("/otp/send", otpEmailHandler.SendOTP)
		emailGroup.POST("/otp/verify", otpEmailHandler.VerifyOTP)
		emailGroup.POST("/magic-link/send", otpEmailHandler.SendMagicLink)
		emailGroup.GET("/magic-link/verify", otpEmailHandler.VerifyMagicLink)
		emailGroup.POST("/send", otpEmailHandler.SendEmail)
	}

	pushGroup := r.Group("/api/v1/push")
	{
		pushGroup.POST("/send", pushHandler.SendPush)
		pushGroup.POST("/device/register", pushHandler.RegisterDevice)
		pushGroup.GET("/user/:user_id/devices", pushHandler.GetUserDevices)
		pushGroup.DELETE("/device/:user_id/:device_token", deviceHandler.UnregisterDevice)
	}

	wechatGroup := r.Group("/api/v1/wechat")
	{
		wechatGroup.POST("/template/send", wechatTemplateHandler.SendTemplate)
	}

	messageGroup := r.Group("/api/v1/messages")
	{
		messageGroup.POST("", messageHandler.CreateMessage)
		messageGroup.GET("", messageHandler.ListMessages)
		messageGroup.PUT("/:id/read", messageHandler.MarkRead)
		messageGroup.POST("/read-all", messageHandler.MarkAllRead)
	}

	templateGroup := r.Group("/api/v1/templates")
	{
		templateGroup.POST("", templateHandler.CreateTemplate)
		templateGroup.GET("", templateHandler.ListTemplates)
		templateGroup.GET("/:id", templateHandler.GetTemplate)
		templateGroup.PUT("/:id", templateHandler.UpdateTemplate)
		templateGroup.DELETE("/:id", templateHandler.DeleteTemplate)
		templateGroup.GET("/:id/records", templateHandler.ListSendRecords)
	}

	port := getEnv("PORT", "30311")

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

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
