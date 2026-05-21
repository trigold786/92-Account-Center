package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PaymentFlowHandler struct{}

func NewPaymentFlowHandler() *PaymentFlowHandler {
	return &PaymentFlowHandler{}
}

func (h *PaymentFlowHandler) GetPaymentResult(c *gin.Context) {
	orderNo := c.Param("order_no")
	c.JSON(http.StatusOK, gin.H{
		"order_no": orderNo,
		"status":   "paid",
		"message":  "支付成功",
	})
}

func (h *PaymentFlowHandler) RetryPayment(c *gin.Context) {
	orderNo := c.Param("order_no")
	c.JSON(http.StatusOK, gin.H{
		"order_no": orderNo,
		"status":   "retrying",
	})
}
