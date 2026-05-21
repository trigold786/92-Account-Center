package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/payment-service/internal/service"
)

type RefundHandler struct {
	svc *service.RefundService
}

func NewRefundHandler(svc *service.RefundService) *RefundHandler {
	return &RefundHandler{svc: svc}
}

func (h *RefundHandler) RequestRefund(c *gin.Context) {
	var req struct {
		OrderID int64  `json:"order_id"`
		Reason  string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	refund, err := h.svc.RequestRefund(c.Request.Context(), req.OrderID, userID.(int64), req.Reason)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, refund)
}

func (h *RefundHandler) ApproveRefund(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	adminID, _ := c.Get("user_id")
	if err := h.svc.ApproveRefund(c.Request.Context(), id, adminID.(int64)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "approved"})
}

func (h *RefundHandler) RejectRefund(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Note string `json:"note"`
	}
	c.ShouldBindJSON(&req)
	adminID, _ := c.Get("user_id")
	if err := h.svc.RejectRefund(c.Request.Context(), id, adminID.(int64), req.Note); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}
