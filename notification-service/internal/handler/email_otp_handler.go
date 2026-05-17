package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
	"github.com/trigold786/92-Account-Center/notification-service/internal/service"
)

type OTPEmailHandler struct {
	emailService service.OTPEmailService
}

func NewOTPEmailHandler(emailService service.OTPEmailService) *OTPEmailHandler {
	return &OTPEmailHandler{emailService: emailService}
}

func (h *OTPEmailHandler) SendOTP(c *gin.Context) {
	var req model.SendOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request body"})
		return
	}

	resp, err := h.emailService.SendOTP(c.Request.Context(), req.Email)
	if err != nil {
		if err == service.ErrRateLimitExceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{"code": 429, "message": "rate limit exceeded"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "OTP sent successfully",
		"data":    resp,
	})
}

func (h *OTPEmailHandler) VerifyOTP(c *gin.Context) {
	var req model.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request body"})
		return
	}

	success, err := h.emailService.VerifyOTP(c.Request.Context(), req.Email, req.Code)
	if err != nil {
		if err == service.ErrInvalidOTP {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid verification code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "OTP verified",
		"data": gin.H{
			"success": success,
		},
	})
}

func (h *OTPEmailHandler) SendMagicLink(c *gin.Context) {
	var req model.SendMagicLinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request body"})
		return
	}

	resp, err := h.emailService.SendMagicLink(c.Request.Context(), req.Email, req.TargetURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Magic link sent successfully",
		"data":    resp,
	})
}

func (h *OTPEmailHandler) VerifyMagicLink(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "token is required"})
		return
	}

	email, err := h.emailService.VerifyMagicLink(c.Request.Context(), token)
	if err != nil {
		if err == service.ErrInvalidMagicLink {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid magic link"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Magic link verified",
		"data": model.VerifyMagicLinkResponse{
			Email: email,
		},
	})
}

func (h *OTPEmailHandler) SendEmail(c *gin.Context) {
	var req model.SendEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request body"})
		return
	}

	err := h.emailService.SendEmail(c.Request.Context(), req.To, req.Subject, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Email sent successfully",
	})
}
