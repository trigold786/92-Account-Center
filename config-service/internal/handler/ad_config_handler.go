package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/config-service/internal/service"
)

type AdConfigHandler struct {
	svc *service.AdConfigService
}

func NewAdConfigHandler(svc *service.AdConfigService) *AdConfigHandler {
	return &AdConfigHandler{svc: svc}
}

func (h *AdConfigHandler) GetAdConfig(c *gin.Context) {
	level := c.DefaultQuery("level", "L0")
	config, err := h.svc.GetAdConfig(c.Request.Context(), level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, config)
}

func (h *AdConfigHandler) CheckFrequency(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	placement := c.Query("placement")
	allowed, err := h.svc.CheckFrequencyControl(c.Request.Context(), userID.(string), placement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if !allowed {
		c.JSON(http.StatusTooManyRequests, gin.H{"allowed": false, "message": "frequency limit exceeded"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"allowed": true})
}
