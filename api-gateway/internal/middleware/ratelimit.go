package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type TokenBucket struct {
	Tokens   float64
	LastRef  time.Time
	LastUsed time.Time
	Mu       sync.Mutex
}

type IPRateLimiter struct {
	Buckets sync.Map
	RPS     float64
}

func NewIPRateLimiter(rps int) *IPRateLimiter {
	l := &IPRateLimiter{RPS: float64(rps)}
	go l.evictStale()
	return l
}

func (l *IPRateLimiter) evictStale() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		cutoff := time.Now().Add(-10 * time.Minute)
		l.Buckets.Range(func(key, value any) bool {
			b := value.(*TokenBucket)
			b.Mu.Lock()
			stale := b.LastUsed.Before(cutoff)
			b.Mu.Unlock()
			if stale {
				l.Buckets.Delete(key)
			}
			return true
		})
	}
}

func (l *IPRateLimiter) GetBucket(ip string) *TokenBucket {
	now := time.Now()
	val, _ := l.Buckets.LoadOrStore(ip, &TokenBucket{
		Tokens:   l.RPS,
		LastRef:  now,
		LastUsed: now,
	})
	b := val.(*TokenBucket)
	b.Mu.Lock()
	b.LastUsed = now
	b.Mu.Unlock()
	return b
}

func (l *IPRateLimiter) Allow(ip string) bool {
	b := l.GetBucket(ip)
	b.Mu.Lock()
	defer b.Mu.Unlock()
	now := time.Now()
	elapsed := now.Sub(b.LastRef).Seconds()
	b.Tokens += elapsed * l.RPS
	if b.Tokens > l.RPS {
		b.Tokens = l.RPS
	}
	if b.Tokens < 1 {
		b.LastRef = now
		return false
	}
	b.Tokens--
	b.LastRef = now
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
