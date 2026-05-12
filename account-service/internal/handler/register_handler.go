package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type RegisterRequest struct {
	PhoneNumber  string `json:"phone_number" binding:"required"`
	AccountID    string `json:"account_id" binding:"required"`
	Password     string `json:"password" binding:"required"`
	AgreeToTerms bool   `json:"agree_to_terms" binding:"required"`
}

type RegisterResponse struct {
	ID          int64  `json:"id"`
	PhoneNumber string `json:"phone_number"`
	AccountID   string `json:"account_id"`
	Message     string `json:"message"`
}

type RegisterHandler struct {
	userService service.UserService
}

func NewRegisterHandler(userService service.UserService) *RegisterHandler {
	return &RegisterHandler{userService: userService}
}

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
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		AccountID:   user.AccountID,
		Message:     "User registered successfully",
	})
}
