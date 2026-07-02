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
	_, _ = rand.Read(state)
	stateStr := hex.EncodeToString(state)

	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("oauth_state", stateStr, 300, "/", "", false, true)

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
	state := c.Query("state")

	if provider == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider and code are required"})
		return
	}

	stateCookie, _ := c.Cookie("oauth_state")
	if stateCookie == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing oauth state"})
		return
	}
	if state == "" || stateCookie != state {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}
	c.SetCookie("oauth_state", "", -1, "/", "", false, true)

	result, err := h.oauthService.HandleCallback(c.Request.Context(), provider, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}
