package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sunxi/92-Account-Center/account-service/internal/service"
)

// RegisterRequest represents the request body for user registration.
type RegisterRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	AccountID   string `json:"account_id" binding:"required"`
	Password    string `json:"password" binding:"required"`
	AgreeToTerms bool  `json:"agree_to_terms" binding:"required"`
}

// RegisterResponse represents the response for user registration.
type RegisterResponse struct {
	ID        int64  `json:"id"`
	PhoneNumber string `json:"phone_number"`
	AccountID   string `json:"account_id"`
	Message     string `json:"message"`
}

// RegisterHandler handles user registration requests.
type RegisterHandler struct {
	userService service.UserService
}

// NewRegisterHandler creates a new RegisterHandler.
func NewRegisterHandler(userService service.UserService) *RegisterHandler {
	return &RegisterHandler{userService: userService}
}

// Register handles the registration endpoint.
func (h *RegisterHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), req.PhoneNumber, req.AccountID, req.Password, req.AgreeToTerms)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, RegisterResponse{
		ID:         user.ID,
		PhoneNumber: user.PhoneNumber,
		AccountID:   user.AccountID,
		Message:     "User registered successfully",
	})
}