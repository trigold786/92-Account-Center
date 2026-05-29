package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type RouteRoleConfig struct {
	Prefix string
	Roles  []string
}

func RoleGuardMiddleware(routes []RouteRoleConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		for _, route := range routes {
			if !strings.HasPrefix(path, route.Prefix) {
				continue
			}

			if len(route.Roles) == 0 {
				break
			}

			roles, exists := c.Get("roles")
			if !exists {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}

			userRoles, ok := roles.([]string)
			if !ok {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}

			allowed := false
			for _, userRole := range userRoles {
				for _, allowedRole := range route.Roles {
					if userRole == allowedRole {
						allowed = true
						break
					}
				}
				if allowed {
					break
				}
			}

			if !allowed {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
				return
			}
			break
		}

		c.Next()
	}
}
