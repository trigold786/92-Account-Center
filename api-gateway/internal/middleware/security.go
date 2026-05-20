package middleware

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func HMACVerifyMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "GET" {
			c.Next()
			return
		}
		timestamp := c.GetHeader("X-Timestamp")
		signature := c.GetHeader("X-Signature")
		if timestamp == "" || signature == "" {
			c.Next()
			return
		}
		body, _ := io.ReadAll(c.Request.Body)
		c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		signPayload := timestamp + ":" + string(body)
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write([]byte(signPayload))
		expected := hex.EncodeToString(mac.Sum(nil))
		if !hmac.Equal([]byte(signature), []byte(expected)) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
			return
		}
		c.Next()
	}
}

func SanitizeInputMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, param := range c.Request.URL.Query() {
			for i, v := range param {
				param[i] = SanitizeString(v)
			}
		}
		c.Next()
	}
}

func SanitizeString(s string) string {
	result := strings.ReplaceAll(s, "<", "&lt;")
	result = strings.ReplaceAll(result, ">", "&gt;")
	result = strings.ReplaceAll(result, "'", "&#39;")
	result = strings.ReplaceAll(result, "\"", "&quot;")
	return result
}
