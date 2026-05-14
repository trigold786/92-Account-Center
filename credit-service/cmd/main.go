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

	"github.com/trigold786/92-Account-Center/credit-service/internal/handler"
	"github.com/trigold786/92-Account-Center/credit-service/internal/repository"
	"github.com/trigold786/92-Account-Center/credit-service/internal/service"
	"github.com/trigold786/92-Account-Center/credit-service/internal/worker"
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

	redisAddr := getEnv("REDIS_URL", "localhost:6379")
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})
	defer rdb.Close()
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Printf("Warning: Redis not available: %v", err)
	}

	creditRepo := repository.NewCreditRepository(db)
	referralRepo := repository.NewReferralRepository(db)

	creditSvc := service.NewCreditService(creditRepo, db)
	referralSvc := service.NewReferralService(referralRepo)
	rebateSvc := service.NewRebateService(creditRepo, referralRepo, creditSvc)

	subWorker := worker.NewSubscriptionWorker(rdb, rebateSvc)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go subWorker.Start(ctx)

	creditHandler := handler.NewCreditHandler(creditSvc)
	referralHandler := handler.NewReferralHandler(referralSvc)

	r := gin.Default()

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

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30312")
	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		cancel()
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("Credit service starting on :%s", port)
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
