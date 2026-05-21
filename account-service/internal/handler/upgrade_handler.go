package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type UpgradeHandler struct {
	svc *service.UpgradeService
}

func NewUpgradeHandler(svc *service.UpgradeService) *UpgradeHandler {
	return &UpgradeHandler{svc: svc}
}

func (h *UpgradeHandler) PreviewUpgrade(c *gin.Context) {
	var req struct {
		CurrentPlan string `json:"current_plan"`
		TargetPlan  string `json:"target_plan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	preview, err := h.svc.PreviewUpgrade(c.Request.Context(), userID.(int64), req.CurrentPlan, req.TargetPlan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (h *UpgradeHandler) PreviewDowngrade(c *gin.Context) {
	var req struct {
		CurrentPlan string `json:"current_plan"`
		TargetPlan  string `json:"target_plan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	preview, err := h.svc.PreviewDowngrade(c.Request.Context(), userID.(int64), req.CurrentPlan, req.TargetPlan)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (h *UpgradeHandler) ExecuteUpgrade(c *gin.Context) {
	var req struct {
		TargetPlan string `json:"target_plan"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	if err := h.svc.ExecuteUpgrade(c.Request.Context(), userID.(int64), req.TargetPlan); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "upgraded", "plan": req.TargetPlan})
}
