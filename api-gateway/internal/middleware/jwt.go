package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/pkg/jwtutil"
)

func JWTAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.URL.Path, "/health") {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		tokenString := jwtutil.ExtractBearerToken(authHeader)
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization token"})
			c.Abort()
			return
		}

		claims, err := jwtutil.ValidateAccessToken(tokenString, jwtSecret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("account_id", claims.AccountID)
		if len(claims.Roles) > 0 {
			c.Set("roles", claims.Roles)
		}

		c.Next()
	}
}
