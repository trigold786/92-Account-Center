package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"account-center/device-fingerprint-service/internal/handler"
	"account-center/device-fingerprint-service/internal/service"
)

func main() {
	deviceService := service.NewDeviceService()
	deviceHandler := handler.NewDeviceHandler(deviceService)

	r := gin.Default()

	deviceGroup := r.Group("/api/v1/device")
	{
		deviceGroup.POST("/register", deviceHandler.RegisterDevice)
		deviceGroup.POST("/verify", deviceHandler.VerifyDevice)
		deviceGroup.POST("/trust", deviceHandler.TrustDevice)
		deviceGroup.GET("/user/:user_id", deviceHandler.GetUserDevices)
		deviceGroup.DELETE("/:device_id", deviceHandler.RemoveDevice)
	}

	port := getEnv("PORT", "8089")
	log.Printf("Device fingerprint service starting on :%s", port)

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