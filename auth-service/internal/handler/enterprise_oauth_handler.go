package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/auth-service/internal/service"
)

type EnterpriseOAuthHandler struct {
	svc *service.EnterpriseOAuthService
}

func NewEnterpriseOAuthHandler(svc *service.EnterpriseOAuthService) *EnterpriseOAuthHandler {
	return &EnterpriseOAuthHandler{svc: svc}
}

func (h *EnterpriseOAuthHandler) WorkWeChatAuth(c *gin.Context) {
	state := generateState()
	url, err := h.svc.GetAuthURL(c.Request.Context(), "work_wechat", state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"auth_url": url, "state": state})
}

func (h *EnterpriseOAuthHandler) DingTalkAuth(c *gin.Context) {
	state := generateState()
	url, err := h.svc.GetAuthURL(c.Request.Context(), "dingtalk", state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"auth_url": url, "state": state})
}

func (h *EnterpriseOAuthHandler) Callback(c *gin.Context) {
	provider := c.Query("provider")
	code := c.Query("code")
	if provider == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider and code are required"})
		return
	}

	token, info, err := h.svc.HandleCallback(c.Request.Context(), provider, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"access_token": token,
		"user_info":    info,
	})
}

func generateState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
