package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sunxi/92-Account-Center/account-service/internal/model"
	"github.com/sunxi/92-Account-Center/account-service/internal/service"
)

// PasswordHandler handles password change requests.
type PasswordHandler struct {
	userService service.UserService
}

// NewPasswordHandler creates a new PasswordHandler.
func NewPasswordHandler(userService service.UserService) *PasswordHandler {
	return &PasswordHandler{userService: userService}
}

// SendVerificationCode handles sending verification code for password change.
func (h *PasswordHandler) SendVerificationCode(c *gin.Context) {
	var req model.SendVerificationCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.userService.SendPasswordVerificationCode(c.Request.Context(), req.ContactType, req.ContactValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Verification code sent successfully"})
}

// ChangePassword handles password change requests.
func (h *PasswordHandler) ChangePassword(c *gin.Context) {
	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	if err := h.userService.ChangePassword(c.Request.Context(), userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}