package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type TierRepository interface {
	GetIdentityTier(ctx context.Context, userID int64) (int, error)
	UpdateIdentityTier(ctx context.Context, userID int64, tier int) error
}

type TierHandler struct {
	repo TierRepository
}

func NewTierHandler(repo TierRepository) *TierHandler {
	return &TierHandler{repo: repo}
}

func (h *TierHandler) GetTier(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	tier, err := h.repo.GetIdentityTier(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":      userID,
		"identity_tier": tier,
	})
}

func (h *TierHandler) UpdateTier(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	var req struct {
		Tier int `json:"tier" binding:"required,min=0,max=4"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.UpdateIdentityTier(c.Request.Context(), userID, req.Tier); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":      userID,
		"identity_tier": req.Tier,
	})
}
