package main

import (
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/lib/pq"

	"github.com/trigold786/92-Account-Center/account-service/internal/cache"
	"github.com/trigold786/92-Account-Center/account-service/internal/handler"
	"github.com/trigold786/92-Account-Center/account-service/internal/repository"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
	"github.com/trigold786/92-Account-Center/account-service/pkg/sms"
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

	redisURL := getEnv("REDIS_URL", "redis://localhost:6379")
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatalf("Failed to parse REDIS_URL: %v", err)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()

	log.Println("Connected to Redis")

	userRepo := repository.NewUserRepository(db)
	entitlementRepo := repository.NewEntitlementRepository(db)
	subscriptionRepo := repository.NewSubscriptionRepository(db)
	smsClient := sms.NewClient(getEnv("SMS_SERVICE_URL", "http://localhost:8083"))

	entitlementCache := cache.NewEntitlementCache(rdb)

	userService := service.NewUserService(userRepo, smsClient)
	entitlementService := service.NewEntitlementService(entitlementRepo, entitlementCache)
	subscriptionService := service.NewSubscriptionService(subscriptionRepo, userRepo, entitlementService)

	registerHandler := handler.NewRegisterHandler(userService)
	passwordHandler := handler.NewPasswordHandler(userService)
	deletionHandler := handler.NewDeletionHandler(userService)
	tierHandler := handler.NewTierHandler(userRepo)
	entitlementHandler := handler.NewEntitlementHandler(entitlementService)
	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionService)

	r := gin.Default()

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

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30301")
	log.Printf("Account service starting on :%s", port)

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
