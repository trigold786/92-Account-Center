package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/notification-service/internal/service"
)

type SMSHandler struct {
	smsService service.SMSService
}

func NewSMSHandler(smsService service.SMSService) *SMSHandler {
	return &SMSHandler{smsService: smsService}
}

func (h *SMSHandler) SendSMS(c *gin.Context) {
	var req struct {
		PhoneNumber  string            `json:"phone_number" binding:"required"`
		TemplateCode string            `json:"template_code"`
		Params       map[string]string `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}

	if err := h.smsService.SendCode(c.Request.Context(), req.PhoneNumber); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "验证码发送成功"})
}

func (h *SMSHandler) GetProviderStatus(c *gin.Context) {
	statuses := h.smsService.GetProviderStatus()

	providers := make([]map[string]string, len(statuses))
	for i, s := range statuses {
		providers[i] = map[string]string{"name": s.Name, "status": s.Status}
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": gin.H{"providers": providers},
	})
}

func (h *SMSHandler) VerifyCode(c *gin.Context) {
	var req struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
		Code        string `json:"code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request"})
		return
	}

	valid, err := h.smsService.VerifyCode(c.Request.Context(), req.PhoneNumber, req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	if !valid {
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"valid": false}})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"valid": true}})
}
