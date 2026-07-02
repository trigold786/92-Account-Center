package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/pkg/jwtutil"
)

var adminRoles = map[string]bool{
	"admin":        true,
	"system_owner": true,
	"operator":     true,
	"finance":      true,
}

func AdminAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		hasAdminRole := false
		for _, role := range claims.Roles {
			if adminRoles[role] {
				hasAdminRole = true
				break
			}
		}
		if !hasAdminRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("account_id", claims.AccountID)
		c.Set("roles", claims.Roles)

		c.Next()
	}
}
