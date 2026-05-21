package handler

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type RegisterRequest struct {
	PhoneNumber  string `json:"phone_number" binding:"required"`
	AccountID    string `json:"account_id" binding:"required"`
	Password     string `json:"password" binding:"required"`
	AgreeToTerms bool   `json:"agree_to_terms" binding:"required"`
	ReferralCode string `json:"referral_code,omitempty"`
}

type RegisterResponse struct {
	ID          int64  `json:"id"`
	PhoneNumber string `json:"phone_number"`
	AccountID   string `json:"account_id"`
	Message     string `json:"message"`
}

type ReferralBinder interface {
	BindReferral(ctx context.Context, referralCode, refereeID string) error
}

type RegisterHandler struct {
	userService    service.UserService
	referralBinder ReferralBinder
}

func NewRegisterHandler(userService service.UserService, referralBinder ReferralBinder) *RegisterHandler {
	return &RegisterHandler{userService: userService, referralBinder: referralBinder}
}

func (h *RegisterHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	user, err := h.userService.Register(c.Request.Context(), req.PhoneNumber, req.AccountID, req.Password, req.AgreeToTerms)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "internal error"})
		return
	}

	if req.ReferralCode != "" && h.referralBinder != nil {
		go func() {
			if err := h.referralBinder.BindReferral(context.Background(), req.ReferralCode, fmt.Sprintf("%d", user.ID)); err != nil {
				slog.Warn("failed to bind referral", "referral_code", req.ReferralCode, "user_id", user.ID, "error", err)
			}
		}()
	}

	c.JSON(http.StatusCreated, RegisterResponse{
		ID:          user.ID,
		PhoneNumber: user.PhoneNumber,
		AccountID:   user.AccountID,
		Message:     "User registered successfully",
	})
}
