package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type DashboardHandler struct {
	svc *service.DashboardService
}

func NewDashboardHandler(svc *service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	userID, _ := c.Get("user_id")
	level := c.DefaultQuery("level", "L0")
	dash, err := h.svc.GetDashboard(c.Request.Context(), userID.(int64), level)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, dash)
}
