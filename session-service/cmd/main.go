package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	"account-center/session-service/internal/handler"
	"account-center/session-service/internal/repository"
	"account-center/session-service/internal/service"
)

func main() {
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPassword := getEnv("REDIS_PASSWORD", "")
	redisDB := getEnv("REDIS_DB", "0")

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       redisDB,
	})
	defer redisClient.Close()

	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")

	maxSessions := getEnvInt("MAX_CONCURRENT_SESSIONS", 5)

	sessionRepo := repository.NewSessionRepository(redisClient)
	sessionService := service.NewSessionService(sessionRepo, int64(maxSessions))
	sessionHandler := handler.NewSessionHandler(sessionService)

	r := gin.Default()

	sessionGroup := r.Group("/session")
	{
		sessionGroup.POST("/create", sessionHandler.CreateSession)
		sessionGroup.POST("/validate", sessionHandler.ValidateSession)
		sessionGroup.GET("/user/:user_id", sessionHandler.GetUserSessions)
		sessionGroup.POST("/invalidate", sessionHandler.InvalidateSession)
		sessionGroup.POST("/invalidate-all", sessionHandler.InvalidateAllUserSessions)
		sessionGroup.POST("/refresh", sessionHandler.RefreshSession)
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down...")
		os.Exit(0)
	}()

	port := getEnv("PORT", "8080")
	log.Printf("Session service starting on :%s", port)
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

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		for _, c := range value {
			if c < '0' || c > '9' {
				return defaultValue
			}
			result = result*10 + int(c-'0')
		}
		return result
	}
	return defaultValue
}
