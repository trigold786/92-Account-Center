package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sunxi/92-Account-Center/auth-service/internal/model"
	"github.com/sunxi/92-Account-Center/auth-service/internal/service"
)

// LoginRequest represents the request body for login.
type LoginRequest struct {
	Credential string `json:"credential" binding:"required"`
	Password   string `json:"password,omitempty"`
	Code       string `json:"code,omitempty"` // For SMS/email OTP
	MagicLink  string `json:"magic_link,omitempty"` // For magic link login
}

// LoginResponse represents the response for login.
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       int64  `json:"user_id"`
	AccountID    string `json:"account_id"`
}

// LoginHandler handles login requests.
type LoginHandler struct {
	authService service.AuthService
}

// NewLoginHandler creates a new LoginHandler.
func NewLoginHandler(authService service.AuthService) *LoginHandler {
	return &LoginHandler{authService: authService}
}

// Login handles the login endpoint.
func (h *LoginHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &model.LoginRequest{
		Credential: req.Credential,
		Password:   req.Password,
		Code:       req.Code,
		MagicLink:  req.MagicLink,
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		UserID:       resp.UserID,
		AccountID:    resp.AccountID,
	})
}

// RefreshToken handles the token refresh endpoint.
func (h *LoginHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.RefreshToken(c.Request.Context(), req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		ExpiresIn:    resp.ExpiresIn,
		UserID:       resp.UserID,
		AccountID:    resp.AccountID,
	})
}