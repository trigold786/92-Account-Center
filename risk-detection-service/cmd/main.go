package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/trigold786/92-Account-Center/risk-detection-service/internal/handler"
	"github.com/trigold786/92-Account-Center/risk-detection-service/internal/repository"
	"github.com/trigold786/92-Account-Center/risk-detection-service/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
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

	riskRepo := repository.NewRiskRepository(db)
	geoService := service.NewGeoService()
	riskService := service.NewRiskService(riskRepo, geoService)
	riskHandler := handler.NewRiskHandler(riskService)

	r := gin.Default()

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	riskHandler.RegisterRoutes(r)

	port := getEnv("PORT", "30306")
	log.Printf("Risk Detection Service starting on :%s", port)
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