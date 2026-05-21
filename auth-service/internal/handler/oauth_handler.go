package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/auth-service/internal/service"
)

type OAuthHandler struct {
	oauthService *service.OAuthService
}

func NewOAuthHandler(svc *service.OAuthService) *OAuthHandler {
	return &OAuthHandler{oauthService: svc}
}

func (h *OAuthHandler) Authorize(c *gin.Context) {
	provider := c.Query("provider")
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}
	state := make([]byte, 16)
	rand.Read(state)
	stateStr := hex.EncodeToString(state)
	url, err := h.oauthService.GetAuthURL(c.Request.Context(), provider, stateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"auth_url": url, "state": stateStr})
}

func (h *OAuthHandler) Callback(c *gin.Context) {
	provider := c.Query("provider")
	code := c.Query("code")
	if provider == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider and code are required"})
		return
	}
	token, info, err := h.oauthService.HandleCallback(c.Request.Context(), provider, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"user_info":    info,
	})
}
