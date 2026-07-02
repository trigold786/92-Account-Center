package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/pkg/env"
)

// CORSMiddleware returns a handler that sets CORS headers from the allowed origins env var.
func CORSMiddleware() gin.HandlerFunc {
	origins := strings.Split(env.Get("CORS_ALLOWED_ORIGINS", "http://localhost:30317,http://localhost:30318"), ",")
	allowed := make(map[string]bool, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o != "" {
			allowed[o] = true
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowed[origin] {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
