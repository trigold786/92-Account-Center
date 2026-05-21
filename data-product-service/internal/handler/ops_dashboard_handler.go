package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type OpsDashboardHandler struct {
	svc *service.MetricsService
}

func NewOpsDashboardHandler(svc *service.MetricsService) *OpsDashboardHandler {
	return &OpsDashboardHandler{svc: svc}
}

func (h *OpsDashboardHandler) GetRegistrationTrends(c *gin.Context) {
	period := c.DefaultQuery("period", "daily")
	trends, err := h.svc.GetRegistrationTrends(c.Request.Context(), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, trends)
}

func (h *OpsDashboardHandler) GetConversionFunnel(c *gin.Context) {
	funnel, err := h.svc.GetConversionFunnel(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, funnel)
}

func (h *OpsDashboardHandler) GetMRR(c *gin.Context) {
	mrr, err := h.svc.GetMRR(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, mrr)
}

func (h *OpsDashboardHandler) GetKFactor(c *gin.Context) {
	k, err := h.svc.GetKFactor(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"k_factor": k})
}

func (h *OpsDashboardHandler) GetRFM(c *gin.Context) {
	rfm, err := h.svc.GetRFM(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rfm)
}
