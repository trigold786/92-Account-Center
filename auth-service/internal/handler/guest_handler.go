package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
	"github.com/trigold786/92-Account-Center/auth-service/internal/service"
)

type GuestHandler struct {
	guestSvc *service.GuestService
}

func NewGuestHandler(guestSvc *service.GuestService) *GuestHandler {
	return &GuestHandler{guestSvc: guestSvc}
}

func (h *GuestHandler) CreateGuest(c *gin.Context) {
	var req model.CreateGuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	guest, err := h.guestSvc.CreateGuest(c.Request.Context(), req.DeviceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create guest session"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"account_id": guest.AccountID,
		"status":     guest.Status,
	})
}

func (h *GuestHandler) UpgradeGuest(c *gin.Context) {
	var req model.UpgradeGuestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	guest, err := h.guestSvc.UpgradeGuest(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "guest session not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id": guest.AccountID,
		"status":     guest.Status,
		"email":      guest.Email,
		"phone":      guest.Phone,
	})
}
