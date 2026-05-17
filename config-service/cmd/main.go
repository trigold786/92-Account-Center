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

	"github.com/trigold786/92-Account-Center/config-service/internal/handler"
	"github.com/trigold786/92-Account-Center/config-service/internal/repository"
	"github.com/trigold786/92-Account-Center/config-service/internal/service"
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

	// Repositories
	configRepo := repository.NewConfigRepository(db)
	releaseRepo := repository.NewReleaseRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	roleRepo := repository.NewRoleRepository(db)

	// Services
	auditSvc := service.NewAuditService(auditRepo)
	configSvc := service.NewConfigService(configRepo, auditSvc)
	releaseSvc := service.NewReleaseService(releaseRepo, configRepo, auditSvc)
	permSvc := service.NewPermissionService(roleRepo, auditSvc)

	// Handlers
	configH := handler.NewConfigHandler(configSvc)
	releaseH := handler.NewReleaseHandler(releaseSvc)
	auditH := handler.NewAuditHandler(auditSvc)
	permH := handler.NewPermissionHandler(permSvc)

	r := gin.Default()

	// Metrics middleware
	r.Use(func(c *gin.Context) {
		start := time.Now()
		c.Next()
		atomic.AddUint64(&requestCount, 1)
		atomic.AddUint64(&durationSumNano, uint64(time.Since(start).Nanoseconds()))
		atomic.AddUint64(&durationCount, 1)
	})

	// Auth middleware
	authMiddleware := func(requiredPermission string) gin.HandlerFunc {
		return func(c *gin.Context) {
			operator := c.GetHeader("X-Operator")
			if operator == "" {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 4, "message": "X-Operator header required"})
				return
			}
			c.Set("operator", operator)

			if requiredPermission != "" {
				allowed, err := permSvc.CheckPermission(c.Request.Context(), operator, requiredPermission)
				if err != nil || !allowed {
					c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"code": 5, "message": "permission denied: " + requiredPermission})
					return
				}
			}

			ip := c.ClientIP()
			ctx := service.WithClientIP(c.Request.Context(), ip)
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		}
	}

	// API routes - Config Groups
	configGroup := r.Group("/api/v1/config", authMiddleware("config.read"))
	{
		configGroup.GET("/groups", configH.ListGroups)
		configGroup.GET("/groups/:id", configH.GetGroupByID)
		configGroup.POST("/groups", authMiddleware("config.edit"), configH.CreateGroup)
		configGroup.PUT("/groups/:id", authMiddleware("config.edit"), configH.UpdateGroup)
		configGroup.DELETE("/groups/:id", authMiddleware("config.delete"), configH.DeleteGroup)

		configGroup.GET("/items", configH.ListItems)
		configGroup.GET("/items/:id", configH.GetItemByID)
		configGroup.POST("/items", authMiddleware("config.edit"), configH.CreateItem)
		configGroup.PUT("/items/:id", authMiddleware("config.edit"), configH.UpdateItem)
		configGroup.DELETE("/items/:id", authMiddleware("config.delete"), configH.DeleteItem)
		configGroup.POST("/items/:id/reset-default", authMiddleware("config.edit"), configH.ResetItemToDefault)
		configGroup.GET("/items/:id/versions", configH.ListVersions)
	}

	// Release routes
	releaseGroup := r.Group("/api/v1/config/releases", authMiddleware("release.create"))
	{
		releaseGroup.GET("", releaseH.ListReleases)
		releaseGroup.GET("/:id", releaseH.GetReleaseByID)
		releaseGroup.POST("", releaseH.CreateRelease)
		releaseGroup.PUT("/:id/submit", authMiddleware("release.submit"), releaseH.SubmitRelease)
		releaseGroup.PUT("/:id/approve", authMiddleware("release.approve"), releaseH.ApproveRelease)
		releaseGroup.PUT("/:id/reject", authMiddleware("release.reject"), releaseH.RejectRelease)
		releaseGroup.POST("/:id/execute", authMiddleware("release.execute"), releaseH.ExecuteRelease)
		releaseGroup.GET("/:id/items", releaseH.ListReleaseItems)
		releaseGroup.POST("/:id/items", releaseH.AddReleaseItem)
	}

	// Audit routes
	auditGroup := r.Group("/api/v1/config/audit-logs", authMiddleware("audit.view"))
	{
		auditGroup.GET("", auditH.ListLogs)
		auditGroup.GET("/:id", auditH.GetLogByID)
	}

	// Permission routes
	permGroup := r.Group("/api/v1/config", authMiddleware("permission.manage"))
	{
		permGroup.GET("/roles", permH.ListRoles)
		permGroup.POST("/roles", permH.CreateRole)
		permGroup.GET("/roles/:id/permissions", permH.GetRolePermissions)
		permGroup.POST("/roles/:id/permissions", permH.AddRolePermission)
		permGroup.GET("/users/:userId/roles", permH.GetUserRoles)
		permGroup.POST("/users/:userId/roles", permH.SetUserRole)
	}

	// Internal routes for service integration (no auth - service-to-service)
	internalGroup := r.Group("/internal/v1/config")
	{
		internalGroup.GET("/items/:code", configH.GetItemByCode)
	}

	r.Any("/metrics", func(c *gin.Context) {
		var buf bytes.Buffer
		fmt.Fprintf(&buf, "# HELP http_requests_total Total HTTP requests\n")
		fmt.Fprintf(&buf, "# TYPE http_requests_total counter\n")
		fmt.Fprintf(&buf, "http_requests_total{service=\"config-service\"} %d\n", atomic.LoadUint64(&requestCount))
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_sum Total request duration in seconds\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_sum counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_sum{service=\"config-service\"} %f\n", time.Duration(atomic.LoadUint64(&durationSumNano)).Seconds())
		fmt.Fprintf(&buf, "# HELP http_request_duration_seconds_count Total request count for duration\n")
		fmt.Fprintf(&buf, "# TYPE http_request_duration_seconds_count counter\n")
		fmt.Fprintf(&buf, "http_request_duration_seconds_count{service=\"config-service\"} %d\n", atomic.LoadUint64(&durationCount))
		fmt.Fprintf(&buf, "# HELP go_goroutines Number of goroutines\n")
		fmt.Fprintf(&buf, "# TYPE go_goroutines gauge\n")
		fmt.Fprintf(&buf, "go_goroutines{service=\"config-service\"} %d\n", runtime.NumGoroutine())
		c.Data(http.StatusOK, "text/plain; version=0.0.4", buf.Bytes())
	})

	r.Any("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	port := getEnv("PORT", "30315")
	log.Printf("Config service starting on :%s", port)

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down gracefully...")
		db.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(shutdownCtx)
	}()

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
