package handler

import (
	"log/slog"
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
	userID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	uid, ok := userID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	refund, err := h.svc.RequestRefund(c.Request.Context(), req.OrderID, uid, req.Reason)
	if err != nil {
		slog.Error("request refund failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, refund)
}

func (h *RefundHandler) ApproveRefund(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	adminID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	aid, ok := adminID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	if err := h.svc.ApproveRefund(c.Request.Context(), id, aid); err != nil {
		slog.Error("approve refund failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
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
	adminID, ok := c.Get("user_id")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	aid, ok := adminID.(int64)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	if err := h.svc.RejectRefund(c.Request.Context(), id, aid, req.Note); err != nil {
		slog.Error("reject refund failed", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "rejected"})
}
