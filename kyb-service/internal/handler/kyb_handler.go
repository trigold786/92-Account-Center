package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/kyb-service/internal/model"
	"github.com/trigold786/92-Account-Center/kyb-service/internal/service"
)

type KYBHandler struct {
	kybService service.KYBService
}

func NewKYBHandler(kybService service.KYBService) *KYBHandler {
	return &KYBHandler{kybService: kybService}
}

func (h *KYBHandler) SubmitEnterprise(c *gin.Context) {
	var req model.EnterpriseInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	resp, err := h.kybService.SubmitEnterpriseInfo(c.Request.Context(), req.UserID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Enterprise info submitted successfully", "data": resp})
}

func (h *KYBHandler) InitiateMicroPayment(c *gin.Context) {
	var req struct {
		EnterpriseID string `json:"enterprise_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	resp, err := h.kybService.InitiateMicroPayment(c.Request.Context(), req.EnterpriseID)
	if err != nil {
		if err == service.ErrEnterpriseNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		if err == service.ErrMicroPaymentNotPending {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to initiate micro payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Micro payment initiated", "data": resp})
}

func (h *KYBHandler) VerifyMicroPayment(c *gin.Context) {
	var req model.MicroPaymentVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	err := h.kybService.VerifyMicroPayment(c.Request.Context(), req.EnterpriseID, req.Amount)
	if err != nil {
		if err == service.ErrEnterpriseNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		if err == service.ErrMicroPaymentNotPending || err == service.ErrInvalidAmount {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to verify micro payment"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Micro payment verified successfully"})
}

func (h *KYBHandler) SubmitFaceVerification(c *gin.Context) {
	var req model.FaceVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	err := h.kybService.SubmitFaceVerification(c.Request.Context(), req.EnterpriseID, req.Token)
	if err != nil {
		if err == service.ErrEnterpriseNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		if err == service.ErrFaceVerificationFailed {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to submit face verification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "Face verification submitted successfully"})
}

func (h *KYBHandler) GetEnterpriseStatus(c *gin.Context) {
	enterpriseID := c.Param("enterprise_id")
	if enterpriseID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "enterprise_id is required"})
		return
	}

	enterpriseID = strings.TrimSpace(enterpriseID)

	resp, err := h.kybService.GetEnterpriseStatus(c.Request.Context(), enterpriseID)
	if err != nil {
		if err == service.ErrEnterpriseNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to get enterprise status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": resp})
}