package middleware

import (
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	PhoneRegex = regexp.MustCompile(`"phone_number"\s*:\s*"(\d{3})\d{4}(\d{4})"`)
	EmailRegex = regexp.MustCompile(`"email"\s*:\s*"([a-zA-Z0-9])[a-zA-Z0-9._%+\-]*@([^"]+)"`)
	IPAddrRegex = regexp.MustCompile(`"ip_address"\s*:\s*"(\d{1,3}\.)\d{1,3}\.\d{1,3}(\.\d{1,3})"`)
)

type ResponseCaptureWriter struct {
	gin.ResponseWriter
	Body []byte
	Code int
}

func (w *ResponseCaptureWriter) Status() int {
	return w.Code
}

func (w *ResponseCaptureWriter) Write(b []byte) (int, error) {
	w.Body = append(w.Body, b...)
	return len(b), nil
}

func (w *ResponseCaptureWriter) WriteHeader(code int) {
	w.Code = code
}

func DesensitizeMiddleware(maxBodySize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if path == "/health" || path == "/metrics" || strings.HasPrefix(path, "/internal/") {
			c.Next()
			return
		}

		captureWriter := &ResponseCaptureWriter{ResponseWriter: c.Writer}
		c.Writer = captureWriter

		c.Next()

		status := captureWriter.Code

		flush := func(data []byte) {
			captureWriter.ResponseWriter.WriteHeader(status)
			captureWriter.ResponseWriter.Write(data)
		}

		if status < 200 || status >= 300 {
			flush(captureWriter.Body)
			return
		}

		contentType := c.Writer.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			flush(captureWriter.Body)
			return
		}

		if int64(len(captureWriter.Body)) > maxBodySize {
			flush(captureWriter.Body)
			return
		}

		accountID, _ := c.Get("user_id")
		if accountIDStr, ok := accountID.(string); ok && strings.HasPrefix(accountIDStr, "admin_") {
			flush(captureWriter.Body)
			return
		}

		body := string(captureWriter.Body)
		masked := PhoneRegex.ReplaceAllString(body, `"phone_number":"$1****$2"`)
		masked = EmailRegex.ReplaceAllString(masked, `"email":"$1***@$2"`)
		masked = IPAddrRegex.ReplaceAllString(masked, `"ip_address":"$1*.*$2"`)

		if masked != body {
			c.Header("X-Desensitized", "true")
			captureWriter.ResponseWriter.Header().Del("Content-Length")
			captureWriter.ResponseWriter.WriteHeader(status)
			captureWriter.ResponseWriter.Write([]byte(masked))
		} else {
			flush(captureWriter.Body)
		}
	}
}
