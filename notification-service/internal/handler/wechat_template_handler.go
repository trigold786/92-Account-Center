package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/notification-service/internal/service"
)

type WeChatTemplateHandler struct {
	svc *service.WeChatTemplateService
}

func NewWeChatTemplateHandler(svc *service.WeChatTemplateService) *WeChatTemplateHandler {
	return &WeChatTemplateHandler{svc: svc}
}

func (h *WeChatTemplateHandler) SendTemplate(c *gin.Context) {
	var req struct {
		OpenID       string            `json:"open_id"`
		TemplateType string            `json:"template_type"`
		Data         map[string]string `json:"data"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SendWithRetry(c.Request.Context(), req.OpenID, req.TemplateType, req.Data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "sent"})
}
