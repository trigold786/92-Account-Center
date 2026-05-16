package main

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
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
)

var (
	requestCount    uint64
	durationSumNano uint64
	durationCount   uint64
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

	dataRepo := repository.NewDataRepository(db)
	rfmSvc := service.NewRFMService(dataRepo)
	dashSvc := service.NewDashboardService(dataRepo, rfmSvc)

	rfmHandler := handler.NewRFMHandler(rfmSvc)
	dashHandler := handler.NewDashboardHandler(dashSvc)
	funnelHandler := handler.NewFunnelHandler(dashSvc)

	r := gin.Default()

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
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30314")
	srv := &http.Server{Addr: ":" + port, Handler: r}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

	log.Printf("Data product service starting on :%s", port)
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
