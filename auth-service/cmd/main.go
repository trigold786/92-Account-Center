package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/v9"

	"github.com/trigold786/92-Account-Center/auth-service/internal/handler"
	"github.com/trigold786/92-Account-Center/auth-service/internal/repository"
	"github.com/trigold786/92-Account-Center/auth-service/internal/service"
	"github.com/trigold786/92-Account-Center/auth-service/pkg/jwt"
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

	accessSecret := getEnv("JWT_ACCESS_SECRET", "access-secret-key-change-in-production")
	refreshSecret := getEnv("JWT_REFRESH_SECRET", "refresh-secret-key-change-in-production")
	jwtMgr := jwt.NewJWTManager(accessSecret, refreshSecret)

	userRepo := repository.NewUserRepository(db)

	rdb := redis.NewClient(&redis.Options{
		Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
	defer rdb.Close()

	authService := service.NewAuthService(userRepo, jwtMgr, rdb)
	loginHandler := handler.NewLoginHandler(authService)

	maxSessions := getEnvInt("MAX_CONCURRENT_SESSIONS", 5)
	sessionRepo := repository.NewSessionRepository(rdb)
	sessionSvc := service.NewSessionService(sessionRepo, int64(maxSessions))
	sessionHandler := handler.NewSessionHandler(sessionSvc)

	deviceRepo := repository.NewDeviceRepository(db)
	deviceSvc := service.NewDeviceFingerprintService(deviceRepo, getEnvInt("TRUST_DAYS", 3), 0.3)
	deviceHandler := handler.NewDeviceHandler(deviceSvc)

	qrcodeSvc := service.NewQRCodeService(rdb, jwtMgr)
	qrcodeHandler := handler.NewQRCodeHandler(qrcodeSvc)

	r := gin.Default()

	authGroup := r.Group("/api/v1/auth")
	{
		authGroup.POST("/login", loginHandler.Login)
		authGroup.POST("/refresh", loginHandler.RefreshToken)
		authGroup.POST("/logout", loginHandler.Logout)
		authGroup.POST("/biometric/register", loginHandler.RegisterBiometric)
		authGroup.POST("/biometric/login", loginHandler.LoginWithBiometric)
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

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30302")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("Auth service starting on :%s", port)
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

func getEnvInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	n := 0
	for _, c := range v {
		if c < '0' || c > '9' {
			return defaultValue
		}
		n = n*10 + int(c-'0')
	}
	return n
}
