package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
	"github.com/trigold786/92-Account-Center/credit-service/internal/service"
)

type ReferralHandler struct {
	referralService service.ReferralService
}

func NewReferralHandler(referralService service.ReferralService) *ReferralHandler {
	return &ReferralHandler{referralService: referralService}
}

func (h *ReferralHandler) BindReferral(c *gin.Context) {
	var req model.ReferralBindRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	if err := h.referralService.BindReferral(c.Request.Context(), req.ReferrerCode, req.RefereeID); err != nil {
		if err == service.ErrInvalidReferralCode || err == service.ErrAlreadyReferred || err == service.ErrSelfReferral {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

func (h *ReferralHandler) GenerateLink(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user_id"})
		return
	}

	result, err := h.referralService.GenerateLink(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}

func (h *ReferralHandler) GetSummary(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user_id"})
		return
	}

	result, err := h.referralService.GetSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}
