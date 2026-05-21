package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type AdEventHandler struct {
	svc *service.AdEventService
}

func NewAdEventHandler(svc *service.AdEventService) *AdEventHandler {
	return &AdEventHandler{svc: svc}
}

func (h *AdEventHandler) TrackAdEvent(c *gin.Context) {
	var req struct {
		UserID    int64  `json:"user_id" binding:"required"`
		EventType string `json:"event_type" binding:"required"`
		AdID      string `json:"ad_id"`
		Placement string `json:"placement"`
		Duration  int    `json:"duration"`
		Converted bool   `json:"converted"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	e, err := h.svc.TrackAdEvent(c.Request.Context(), &model.AdEvent{
		UserID:    req.UserID,
		EventType: req.EventType,
		AdID:      req.AdID,
		Placement: req.Placement,
		Duration:  req.Duration,
		Converted: req.Converted,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, e)
}

func (h *AdEventHandler) GetAdMetrics(c *gin.Context) {
	eventType := c.Query("event_type")
	placement := c.Query("placement")

	metrics, err := h.svc.GetAdMetrics(c.Request.Context(), eventType, placement)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"metrics": metrics})
}
