package handler

import (
	"net/http"
	"strings"

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

type LoginHandler struct {
	authService  service.AuthService
	rateLimiter  *service.RateLimiter
}

func NewLoginHandler(authService service.AuthService, rateLimiter *service.RateLimiter) *LoginHandler {
	return &LoginHandler{
		authService: authService,
		rateLimiter: rateLimiter,
	}
}

func (h *LoginHandler) Login(c *gin.Context) {
	ip := c.ClientIP()
	if h.rateLimiter.IsLocked(c.Request.Context(), "login:"+ip) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
		return
	}
	h.rateLimiter.RecordAttempt(c.Request.Context(), "login:"+ip)

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

	rateLimitKey := "biometric:" + req.DeviceFingerprint
	if h.rateLimiter.IsLocked(c.Request.Context(), rateLimitKey) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
		return
	}
	h.rateLimiter.RecordAttempt(c.Request.Context(), rateLimitKey)

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
