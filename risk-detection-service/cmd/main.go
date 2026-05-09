package main

import (
	"database/sql"
	"log"

	"risk-detection-service/internal/handler"
	"risk-detection-service/internal/repository"
	"risk-detection-service/internal/service"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	db, err := sql.Open("postgres", "postgres://localhost:5432/account_center?sslmode=disable")
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

	riskHandler.RegisterRoutes(r)

	log.Println("Risk Detection Service starting on :8085")
	if err := r.Run(":8085"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}