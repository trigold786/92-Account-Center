package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type TokenBucket struct {
	Tokens  float64
	LastRef time.Time
	Mu      sync.Mutex
}

type IPRateLimiter struct {
	Buckets sync.Map
	RPS     float64
}

func NewIPRateLimiter(rps int) *IPRateLimiter {
	return &IPRateLimiter{RPS: float64(rps)}
}

func (l *IPRateLimiter) GetBucket(ip string) *TokenBucket {
	val, _ := l.Buckets.LoadOrStore(ip, &TokenBucket{
		Tokens:  l.RPS,
		LastRef: time.Now(),
	})
	return val.(*TokenBucket)
}

func (l *IPRateLimiter) Allow(ip string) bool {
	b := l.GetBucket(ip)
	b.Mu.Lock()
	defer b.Mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.LastRef).Seconds()
	b.LastRef = now
	b.Tokens += elapsed * l.RPS
	if b.Tokens > l.RPS {
		b.Tokens = l.RPS
	}
	if b.Tokens < 1 {
		return false
	}
	b.Tokens--
	return true
}

func RateLimitMiddleware(rps int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rps)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
			return
		}
		c.Next()
	}
}
