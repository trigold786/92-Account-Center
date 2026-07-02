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
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/auth-service/internal/handler"
	"github.com/trigold786/92-Account-Center/auth-service/internal/provider"
	"github.com/trigold786/92-Account-Center/auth-service/internal/repository"
	"github.com/trigold786/92-Account-Center/auth-service/internal/service"
	"github.com/trigold786/92-Account-Center/auth-service/internal/svcconfig"
	"github.com/trigold786/92-Account-Center/auth-service/pkg/jwt"
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

var logger = logging.NewLogger("auth-service")

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

	accessSecret := env.GetSecret("JWT_ACCESS_SECRET", "")
	if accessSecret == "" {
		logger.Error("JWT_ACCESS_SECRET not configured")
		os.Exit(1)
	}
	refreshSecret := env.GetSecret("JWT_REFRESH_SECRET", "")
	if refreshSecret == "" {
		logger.Error("JWT_REFRESH_SECRET not configured")
		os.Exit(1)
	}

	configSvcURL := env.Get("CONFIG_SERVICE_URL", "http://localhost:30315")
	configClient := config.NewClient(configSvcURL)
	authCfg, err := svcconfig.Load(configClient)
	if err != nil {
	logger.Warn("config-service unavailable, continuing with env/defaults", "error", err)
}
	logger.Info("config loaded",
		"jwt_access", authCfg.JwtAccessTokenExpire,
		"jwt_refresh", authCfg.JwtRefreshTokenExpire,
		"max_attempts", authCfg.LoginMaxAttempts,
		"lockout", authCfg.LoginLockoutDuration,
		"session_timeout", authCfg.SessionTimeout,
		"session_max", authCfg.SessionMaxPerUser,
		"trust_days", authCfg.DeviceTrustDays,
		"qrcode_ttl", authCfg.QRCodeExpire,
	)

	jwtMgr := jwt.NewJWTManager(accessSecret, refreshSecret, authCfg.JwtAccessTokenExpire, authCfg.JwtRefreshTokenExpire)

	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	rdb := redis.NewClient(&redis.Options{
		Addr:     env.Get("REDIS_ADDR", "localhost:6379"),
		Password: env.GetSecret("REDIS_PASSWORD", ""),
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

	authService := service.NewAuthService(userRepo, roleRepo, jwtMgr, rdb, authCfg)
	rateLimiter := service.NewRateLimiter(rdb, authCfg.LoginRateLimitPerIP, time.Minute)
	loginHandler := handler.NewLoginHandler(authService, rateLimiter)

	sessionRepo := repository.NewSessionRepository(rdb, authCfg.SessionTimeout)
	sessionSvc := service.NewSessionService(sessionRepo, int64(authCfg.SessionMaxPerUser), authCfg.SessionTimeout)
	sessionHandler := handler.NewSessionHandler(sessionSvc)

	deviceRepo := repository.NewDeviceRepository(db)
	deviceSvc := service.NewDeviceFingerprintService(deviceRepo, authCfg.DeviceTrustDays, 0.3)
	deviceHandler := handler.NewDeviceHandler(deviceSvc)

	qrcodeSvc := service.NewQRCodeService(rdb, jwtMgr, roleRepo, authCfg.QRCodeExpire)
	qrcodeHandler := handler.NewQRCodeHandler(qrcodeSvc)

	oauthRegistry := service.NewOAuthProviderRegistry()
	oauthRegistry.Register(service.NewWeChatOAuthProvider(
		env.GetSecret("WECHAT_APP_ID", authCfg.WeChatAppID),
		env.GetSecret("WECHAT_SECRET", authCfg.WeChatSecret),
		env.Get("OAUTH_REDIRECT_URI", authCfg.OAuthRedirectURI),
	))
	oauthRegistry.Register(service.NewAppleOAuthProvider(
		env.GetSecret("APPLE_CLIENT_ID", authCfg.AppleClientID),
		env.GetSecret("APPLE_TEAM_ID", authCfg.AppleTeamID),
		env.GetSecret("APPLE_KEY_ID", authCfg.AppleKeyID),
		env.Get("OAUTH_REDIRECT_URI", authCfg.OAuthRedirectURI),
	))
	oauthRegistry.Register(service.NewGoogleOAuthProvider(
		env.GetSecret("GOOGLE_CLIENT_ID", authCfg.GoogleClientID),
		env.GetSecret("GOOGLE_SECRET", authCfg.GoogleSecret),
		env.Get("OAUTH_REDIRECT_URI", authCfg.OAuthRedirectURI),
	))
	oauthRegistry.Register(service.NewAlipayOAuthProvider(
		env.GetSecret("ALIPAY_APP_ID", ""),
		env.GetSecret("ALIPAY_PRIVATE_KEY", ""),
		env.Get("OAUTH_REDIRECT_URI", ""),
	))
	oauthSvc := service.NewOAuthService(oauthRegistry, userRepo, roleRepo, jwtMgr)
	oauthH := handler.NewOAuthHandler(oauthSvc)

	guestSvc := service.NewGuestService(nil)
	guestHandler := handler.NewGuestHandler(guestSvc)

	entOAuthSvc := service.NewEnterpriseOAuthService()
	entOAuthSvc.Register(provider.NewWorkWeChatProvider(
		env.GetSecret("WORK_WECHAT_CORP_ID", ""),
		env.Get("WORK_WECHAT_AGENT_ID", ""),
		env.GetSecret("WORK_WECHAT_CORP_SECRET", ""),
		env.Get("ENTERPRISE_OAUTH_REDIRECT_URI", ""),
	))
	entOAuthSvc.Register(provider.NewDingTalkProvider(
		env.GetSecret("DINGTALK_APP_KEY", ""),
		env.GetSecret("DINGTALK_APP_SECRET", ""),
		env.Get("ENTERPRISE_OAUTH_REDIRECT_URI", ""),
	))
	entOAuthHandler := handler.NewEnterpriseOAuthHandler(entOAuthSvc)

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

	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/login", loginHandler.Login)
		authGroup.POST("/refresh", loginHandler.RefreshToken)
		authGroup.POST("/logout", loginHandler.Logout)
		authGroup.POST("/biometric/register", loginHandler.RegisterBiometric)
		authGroup.POST("/biometric/login", loginHandler.LoginWithBiometric)
		authGroup.GET("/oauth/authorize", oauthH.Authorize)
		authGroup.POST("/oauth/callback", oauthH.Callback)
		authGroup.POST("/guest", guestHandler.CreateGuest)
		authGroup.POST("/guest/upgrade", guestHandler.UpgradeGuest)
	}

	enterpriseGroup := r.Group("/api/v1/auth/enterprise")
	{
		enterpriseGroup.GET("/work-wechat", entOAuthHandler.WorkWeChatAuth)
		enterpriseGroup.GET("/dingtalk", entOAuthHandler.DingTalkAuth)
		enterpriseGroup.GET("/callback", entOAuthHandler.Callback)
	}

	sessionGroup := r.Group("/api/v1/session")
	{
		sessionGroup.POST("/create", sessionHandler.CreateSession)
		sessionGroup.POST("/validate", sessionHandler.ValidateSession)
		sessionGroup.GET("/user/:user_id", sessionHandler.GetUserSessions)
		sessionGroup.POST("/invalidate", sessionHandler.InvalidateSession)
		sessionGroup.POST("/invalidate-all", sessionHandler.InvalidateAllUserSessions)
		sessionGroup.POST("/refresh", sessionHandler.RefreshSession)
	}

	deviceGroup := r.Group("/api/v1/device")
	{
		deviceGroup.POST("/register", deviceHandler.RegisterDevice)
		deviceGroup.POST("/verify", deviceHandler.VerifyDevice)
		deviceGroup.POST("/trust", deviceHandler.TrustDevice)
		deviceGroup.GET("/user/:user_id", deviceHandler.GetUserDevices)
		deviceGroup.DELETE("/:device_id", deviceHandler.RemoveDevice)
	}

	qrcodeGroup := r.Group("/api/v1/qrcode")
	{
		qrcodeGroup.POST("/generate", qrcodeHandler.Generate)
		qrcodeGroup.GET("/:code_id/status", qrcodeHandler.GetStatus)
		qrcodeGroup.POST("/:code_id/scan", qrcodeHandler.Scan)
		qrcodeGroup.POST("/:code_id/confirm", qrcodeHandler.Confirm)
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
		fmt.Fprintf(&buf, "http_requests_total{service=\"auth-service\"} %d\n", atomic.LoadUint64(&requestCount))
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_sum Total request duration in seconds\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_sum counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_sum{service=\"auth-service\"} %f\n", time.Duration(atomic.LoadUint64(&durationSumNano)).Seconds())
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_count Total request count for duration\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_count counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_count{service=\"auth-service\"} %d\n", atomic.LoadUint64(&durationCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"auth-service\"} %d\n", runtime.NumGoroutine())
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

	port := env.Get("PORT", "30302")
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

	logger.Info("starting server", "port", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("failed to start server", "error", err.Error())
		os.Exit(1)
	}
}


