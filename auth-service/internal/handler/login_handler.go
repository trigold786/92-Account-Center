package handler

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
	"github.com/trigold786/92-Account-Center/auth-service/internal/service"
)

type LoginRequest struct {
	Credential          string `json:"credential" binding:"required"`
	Password            string `json:"password,omitempty"`
	Code                string `json:"code,omitempty"`
	MagicLink           string `json:"magic_link,omitempty"`
	DeviceFingerprintID string `json:"device_fingerprint_id,omitempty"`
}

const rateLimitCleanupInterval = 5 * time.Minute
const rateLimitEntryTTL = 2 * time.Minute

type ipRateLimiter struct {
	mu       sync.Mutex
	attempts []time.Time
	lastUsed time.Time
}

func (rl *ipRateLimiter) allow(maxAttempts int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	rl.lastUsed = time.Now()
	now := time.Now()
	cutoff := now.Add(-window)
	var valid []time.Time
	for _, t := range rl.attempts {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.attempts = valid
	if len(valid) >= maxAttempts {
		return false
	}
	rl.attempts = append(rl.attempts, now)
	return true
}

func (rl *ipRateLimiter) isStale() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()
	return time.Since(rl.lastUsed) > rateLimitEntryTTL
}

type LoginHandler struct {
	authService    service.AuthService
	rateLimiters   *sync.Map
	loginRateLimit int
}

func NewLoginHandler(authService service.AuthService, loginRateLimit int) *LoginHandler {
	h := &LoginHandler{
		authService:    authService,
		rateLimiters:   &sync.Map{},
		loginRateLimit: loginRateLimit,
	}
	if h.loginRateLimit <= 0 {
		h.loginRateLimit = 10
	}
	go h.cleanupRateLimiters()
	return h
}

func (h *LoginHandler) cleanupRateLimiters() {
	for {
		time.Sleep(rateLimitCleanupInterval)
		h.rateLimiters.Range(func(key, value interface{}) bool {
			if limiter, ok := value.(*ipRateLimiter); ok && limiter.isStale() {
				h.rateLimiters.Delete(key)
			}
			return true
		})
	}
}

func (h *LoginHandler) Login(c *gin.Context) {
	ip := c.ClientIP()
	limiterIface, _ := h.rateLimiters.LoadOrStore(ip, &ipRateLimiter{})
	limiter := limiterIface.(*ipRateLimiter)
	if !limiter.allow(h.loginRateLimit, time.Minute) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
		return
	}

	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &model.LoginRequest{
		Credential:          req.Credential,
		Password:            req.Password,
		Code:                req.Code,
		MagicLink:           req.MagicLink,
		DeviceFingerprintID: req.DeviceFingerprintID,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, model.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		UserID:       resp.UserID,
		AccountID:    resp.AccountID,
		Roles:        resp.Roles,
	})
}

func (h *LoginHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.JSON(http.StatusOK, model.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		UserID:       resp.UserID,
		AccountID:    resp.AccountID,
		Roles:        resp.Roles,
	})
}

func (h *LoginHandler) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
		return
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header format"})
		return
	}

	accessToken := parts[1]
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "empty token"})
		return
	}

	if err := h.authService.Logout(c.Request.Context(), accessToken); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}

func (h *LoginHandler) RegisterBiometric(c *gin.Context) {
	var req model.BiometricRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.authService.RegisterBiometric(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "生物识别注册成功"})
}

func (h *LoginHandler) LoginWithBiometric(c *gin.Context) {
	var req model.BiometricLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	resp, err := h.authService.LoginWithBiometric(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, model.LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		UserID:       resp.UserID,
		AccountID:    resp.AccountID,
		Roles:        resp.Roles,
	})
}
