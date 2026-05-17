package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type FunnelHandler struct {
	dashSvc service.DashboardService
}

func NewFunnelHandler(dashSvc service.DashboardService) *FunnelHandler {
	return &FunnelHandler{dashSvc: dashSvc}
}

func (h *FunnelHandler) GetSubscriptionFunnel(c *gin.Context) {
	funnel, err := h.dashSvc.GetSubscriptionFunnel(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": funnel})
}
