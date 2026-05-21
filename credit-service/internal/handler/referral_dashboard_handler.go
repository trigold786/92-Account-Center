package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/credit-service/internal/service"
)

type ReferralDashboardHandler struct {
	svc *service.ReferralDashboardService
}

func NewReferralDashboardHandler(svc *service.ReferralDashboardService) *ReferralDashboardHandler {
	return &ReferralDashboardHandler{svc: svc}
}

func (h *ReferralDashboardHandler) GetFunnel(c *gin.Context) {
	userID, _ := c.Get("user_id")
	funnel, err := h.svc.GetFunnel(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, funnel)
}

func (h *ReferralDashboardHandler) GetEarningsTrend(c *gin.Context) {
	userID, _ := c.Get("user_id")
	period := c.DefaultQuery("period", "weekly")
	trend, err := h.svc.GetEarningsTrend(c.Request.Context(), userID.(int64), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trend)
}
