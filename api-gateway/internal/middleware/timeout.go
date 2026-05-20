package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func TimeoutMiddleware(timeoutSec int) gin.HandlerFunc {
	timeout := time.Duration(timeoutSec) * time.Second
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		finished := make(chan struct{})
		go func() {
			defer close(finished)
			c.Next()
		}()

		select {
		case <-finished:
			return
		case <-ctx.Done():
			if !c.IsAborted() {
				c.AbortWithStatusJSON(http.StatusGatewayTimeout, gin.H{
					"error": "request timed out",
				})
			}
		}
	}
}
