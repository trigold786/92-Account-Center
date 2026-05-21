package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func AdminAuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
			return
		}

		tokenStr := parts[1]
		segments := strings.Split(tokenStr, ".")
		if len(segments) != 3 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token format"})
			return
		}

		signingInput := segments[0] + "." + segments[1]
		mac := hmac.New(sha256.New, []byte(jwtSecret))
		mac.Write([]byte(signingInput))
		expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

		if !hmac.Equal([]byte(segments[2]), []byte(expectedSig)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token signature"})
			return
		}

		payloadBytes, err := base64.RawURLEncoding.DecodeString(segments[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token payload"})
			return
		}

		var claims map[string]interface{}
		if err := json.Unmarshal(payloadBytes, &claims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token claims"})
			return
		}

		exp, ok := claims["exp"].(float64)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token missing expiration"})
			return
		}
		if time.Now().Unix() > int64(exp) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
			return
		}

		role, _ := claims["role"].(string)

		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			return
		}

	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		userID, ok = claims["user_id"].(string)
		if !ok || userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing user identity in token"})
			return
		}
	}
	accountID, _ := claims["account_id"].(string)

		c.Set("user_id", userID)
		c.Set("admin_role", role)
		c.Set("account_id", accountID)
		c.Next()
	}
}
