package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/account-service/internal/model"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type EntitlementHandler struct {
	svc service.EntitlementService
}

func NewEntitlementHandler(svc service.EntitlementService) *EntitlementHandler {
	return &EntitlementHandler{svc: svc}
}

func (h *EntitlementHandler) GetUserEntitlements(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	entitlements, err := h.svc.GetUserEntitlements(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entitlements": entitlements})
}

func (h *EntitlementHandler) Consume(c *gin.Context) {
	var req model.ConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, err := service.ParseUserID(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	resp, err := h.svc.ConsumeQuota(c.Request.Context(), userID, req.FeatureCode, req.Amount)
	if err != nil {
		if err == service.ErrInsufficientQuota || err == service.ErrEntitlementNotFound {
			c.JSON(http.StatusConflict, gin.H{"error": "internal error"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *EntitlementHandler) Grant(c *gin.Context) {
	var req model.GrantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, err := service.ParseUserID(req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	if err := h.svc.GrantEntitlements(c.Request.Context(), userID, req.TierLevel); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "entitlements granted"})
}
