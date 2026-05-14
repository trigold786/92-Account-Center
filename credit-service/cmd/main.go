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

	"github.com/trigold786/92-Account-Center/credit-service/internal/handler"
	"github.com/trigold786/92-Account-Center/credit-service/internal/repository"
	"github.com/trigold786/92-Account-Center/credit-service/internal/service"
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

	creditRepo := repository.NewCreditRepository(db)
	referralRepo := repository.NewReferralRepository(db)

	creditService := service.NewCreditService(creditRepo, db)
	referralService := service.NewReferralService(referralRepo)

	creditHandler := handler.NewCreditHandler(creditService)
	referralHandler := handler.NewReferralHandler(referralService)

	r := gin.Default()

	r.GET("/api/v1/credits/:user_id/account", creditHandler.GetAccount)
	r.GET("/api/v1/credits/:user_id/transactions", creditHandler.GetTransactions)

	internalCredits := r.Group("/internal/v1/credits")
	{
		internalCredits.POST("/earn", creditHandler.EarnCredits)
		internalCredits.POST("/consume", creditHandler.ConsumeCredits)
		internalCredits.POST("/refund", creditHandler.RefundCredits)
	}

	r.POST("/api/v1/credits/calculate-discount", creditHandler.CalculateDiscount)

	r.POST("/api/v1/referral/bind", referralHandler.BindReferral)
	r.POST("/api/v1/referral/generate-link", referralHandler.GenerateLink)
	r.GET("/api/v1/referral/:user_id/summary", referralHandler.GetSummary)

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30312")
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server forced shutdown: %v", err)
		}
	}()

	log.Printf("Credit Service starting on :%s", port)
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
