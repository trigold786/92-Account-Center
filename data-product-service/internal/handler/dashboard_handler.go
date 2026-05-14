package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type DashboardHandler struct {
	dashSvc service.DashboardService
}

func NewDashboardHandler(dashSvc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{dashSvc: dashSvc}
}

func (h *DashboardHandler) GetOverview(c *gin.Context) {
	overview, err := h.dashSvc.GetOverview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": overview})
}
