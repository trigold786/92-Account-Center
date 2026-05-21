package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

var deprecatedVersions = map[string]bool{
	"v1": true,
}

var latestVersion = "v2"

func VersionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		version := extractVersion(c.Request.URL.Path)
		if version == "" {
			c.Next()
			return
		}

		c.Set("api_version", version)
		c.Header("X-API-Version", version)

		if deprecatedVersions[version] {
			c.Header("X-Deprecated-Version", fmt.Sprintf("version %s is deprecated, please migrate to %s", version, latestVersion))
			c.Header("Sunset", "Sat, 01 Jan 2027 00:00:00 GMT")
		}

		c.Request.Header.Set("X-API-Version", version)
		c.Next()
	}
}

func extractVersion(path string) string {
	if !strings.HasPrefix(path, "/api/") {
		return ""
	}
	parts := strings.SplitN(path, "/", 4)
	if len(parts) < 3 {
		return ""
	}
	v := parts[2]
	if !strings.HasPrefix(v, "v") {
		return ""
	}
	return v
}
