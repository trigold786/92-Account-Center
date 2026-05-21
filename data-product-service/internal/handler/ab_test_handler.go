package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type ABTestHandler struct {
	svc *service.ABTestService
}

func NewABTestHandler(svc *service.ABTestService) *ABTestHandler {
	return &ABTestHandler{svc: svc}
}

func (h *ABTestHandler) CreateExperiment(c *gin.Context) {
	var req struct {
		Name     string         `json:"name" binding:"required"`
		Variants []model.ABVariant `json:"variants" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	exp, err := h.svc.CreateExperiment(c.Request.Context(), req.Name, req.Variants)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, exp)
}

func (h *ABTestHandler) AssignVariant(c *gin.Context) {
	experimentID := c.Param("id")
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	assignment, err := h.svc.AssignVariant(c.Request.Context(), experimentID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, assignment)
}

func (h *ABTestHandler) RecordEvent(c *gin.Context) {
	experimentID := c.Param("id")
	var req struct {
		UserID    string `json:"user_id" binding:"required"`
		Variant   string `json:"variant" binding:"required"`
		EventType string `json:"event_type" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.RecordEvent(c.Request.Context(), experimentID, req.UserID, req.Variant, req.EventType); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *ABTestHandler) GetResults(c *gin.Context) {
	experimentID := c.Param("id")

	results, err := h.svc.GetResults(c.Request.Context(), experimentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}
