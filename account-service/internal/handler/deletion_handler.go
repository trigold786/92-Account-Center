package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sunxi/92-Account-Center/account-service/internal/model"
	"github.com/sunxi/92-Account-Center/account-service/internal/service"
)

// DeletionRequest represents the request body for account deletion.
type DeletionRequest struct {
	VerificationCode string `json:"verification_code" binding:"required"`
	VerificationType string `json:"verification_type" binding:"required,oneof=sms_code email_otp"`
}

// DeletionResponse represents the response for deletion operations.
type DeletionResponse struct {
	Message string `json:"message"`
}

// DeletionStatusResponse represents the response for deletion status.
type DeletionStatusResponse struct {
	UserID      int64     `json:"user_id"`
	RequestedAt *time.Time `json:"requested_at,omitempty"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	CancelledAt *time.Time `json:"cancelled_at,omitempty"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty"`
}

// DeletionHandler handles account deletion requests.
type DeletionHandler struct {
	userService service.UserService
}

// NewDeletionHandler creates a new DeletionHandler.
func NewDeletionHandler(userService service.UserService) *DeletionHandler {
	return &DeletionHandler{userService: userService}
}

// RequestDeletion handles the account deletion request endpoint.
func (h *DeletionHandler) RequestDeletion(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var req DeletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert userID to string for service call
	userIDStr := userID.(string)

	resp, err := h.userService.RequestAccountDeletion(c.Request.Context(), userIDStr, &model.DeletionRequest{
		UserID:           0, // Will be filled by service
		VerificationCode: req.VerificationCode,
		VerificationType: req.VerificationType,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DeletionResponse{
		Message: resp.Message,
	})
}

// CancelDeletion handles the account deletion cancellation endpoint.
func (h *DeletionHandler) CancelDeletion(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Convert userID to string for service call
	userIDStr := userID.(string)

	resp, err := h.userService.CancelAccountDeletion(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DeletionResponse{
		Message: resp.Message,
	})
}

// GetDeletionStatus handles the account deletion status endpoint.
func (h *DeletionHandler) GetDeletionStatus(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Convert userID to string for service call
	userIDStr := userID.(string)

	deletion, err := h.userService.GetDeletionStatus(c.Request.Context(), userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DeletionStatusResponse{
		UserID:      deletion.UserID,
		RequestedAt: deletion.RequestedAt,
		ExpiresAt:   deletion.ExpiresAt,
		CancelledAt: deletion.CancelledAt,
		DeletedAt:   deletion.DeletedAt,
	})
}