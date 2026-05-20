package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type StreamHandler struct {
	svc *service.StreamService
}

func NewStreamHandler(svc *service.StreamService) *StreamHandler {
	return &StreamHandler{svc: svc}
}

func (h *StreamHandler) ProcessEvent(c *gin.Context) {
	var req struct {
		UserID    int64  `json:"user_id" binding:"required"`
		EventType string `json:"event_type" binding:"required"`
		Payload   string `json:"payload"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	e, err := h.svc.ProcessEvent(c.Request.Context(), &model.StreamEvent{
		UserID:    req.UserID,
		EventType: req.EventType,
		Payload:   req.Payload,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, e)
}

func (h *StreamHandler) GetOnlineCount(c *gin.Context) {
	windowMin, _ := strconv.Atoi(c.DefaultQuery("window_minutes", "5"))
	count, err := h.svc.GetOnlineCount(c.Request.Context(), time.Duration(windowMin)*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"online_count": count})
}

func (h *StreamHandler) GetRealtimeFunnel(c *gin.Context) {
	var req struct {
		Steps []string `json:"steps" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	windowMin, _ := strconv.Atoi(c.DefaultQuery("window_minutes", "5"))
	funnel, err := h.svc.GetRealtimeFunnel(c.Request.Context(), time.Duration(windowMin)*time.Minute, req.Steps)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"funnel": funnel})
}
