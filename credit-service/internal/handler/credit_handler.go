package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/credit-service/internal/model"
	"github.com/trigold786/92-Account-Center/credit-service/internal/service"
)

type CreditHandler struct {
	creditService service.CreditService
}

func NewCreditHandler(creditService service.CreditService) *CreditHandler {
	return &CreditHandler{creditService: creditService}
}

func (h *CreditHandler) GetAccount(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user_id"})
		return
	}

	account, err := h.creditService.GetAccount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": account})
}

func (h *CreditHandler) GetTransactions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user_id"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.creditService.GetTransactions(c.Request.Context(), userID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}

func (h *CreditHandler) EarnCredits(c *gin.Context) {
	var req model.EarnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user_id"})
		return
	}

	if err := h.creditService.EarnCredits(c.Request.Context(), userID, req.Amount, req.Type, req.ReferenceID, req.Details); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

func (h *CreditHandler) ConsumeCredits(c *gin.Context) {
	var req model.ConsumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user_id"})
		return
	}

	if err := h.creditService.ConsumeCredits(c.Request.Context(), userID, req.Amount, req.ReferenceID, req.Details); err != nil {
		if err == service.ErrInsufficientBalance {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		if err == service.ErrAccountFrozen {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

func (h *CreditHandler) RefundCredits(c *gin.Context) {
	var req model.RefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user_id"})
		return
	}

	if err := h.creditService.RefundCredits(c.Request.Context(), userID, req.Amount, req.ReferenceID, req.Details); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success"})
}

func (h *CreditHandler) CalculateDiscount(c *gin.Context) {
	var req model.CalculateDiscountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	userID, err := strconv.ParseInt(req.UserID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user_id"})
		return
	}

	result, err := h.creditService.CalculateDiscount(c.Request.Context(), userID, req.SubscriptionPrice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": result})
}
