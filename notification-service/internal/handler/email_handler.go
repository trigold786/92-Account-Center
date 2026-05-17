package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/notification-service/internal/service"
)

type VerificationEmailHandler struct {
	emailService service.SimpleEmailService
}

func NewVerificationEmailHandler(emailService service.SimpleEmailService) *VerificationEmailHandler {
	return &VerificationEmailHandler{emailService: emailService}
}

func (h *VerificationEmailHandler) SendVerificationCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}

	if err := h.emailService.SendVerificationCode(c.Request.Context(), req.Email); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "failed to send verification code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "验证码发送成功"})
}

func (h *VerificationEmailHandler) VerifyCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}

	valid, err := h.emailService.VerifyCode(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"valid": valid}})
}
