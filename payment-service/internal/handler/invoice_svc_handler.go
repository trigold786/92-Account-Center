package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/service"
)

type InvoiceServiceHandler struct {
	svc *service.InvoiceService
}

func NewInvoiceServiceHandler(svc *service.InvoiceService) *InvoiceServiceHandler {
	return &InvoiceServiceHandler{svc: svc}
}

func (h *InvoiceServiceHandler) CreateInvoiceSvc(c *gin.Context) {
	var req struct {
		OrderID int64   `json:"order_id" binding:"required"`
		Title   string  `json:"title" binding:"required"`
		TaxID   string  `json:"tax_id"`
		Email   string  `json:"email" binding:"required"`
		Amount  float64 `json:"amount"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Get("user_id")
	var userID int64
	switch v := userIDStr.(type) {
	case int64:
		userID = v
	case string:
		userID, _ = strconv.ParseInt(v, 10, 64)
	case float64:
		userID = int64(v)
	}

	inv, err := h.svc.CreateInvoice(c.Request.Context(), &model.Invoice{
		UserID:    userID,
		OrderID:   req.OrderID,
		InvoiceNo: "INV" + strconv.FormatInt(req.OrderID, 10),
		Title:     req.Title,
		TaxID:     req.TaxID,
		Email:     req.Email,
		Amount:    req.Amount,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, inv)
}

func (h *InvoiceServiceHandler) GetInvoiceSvc(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	inv, err := h.svc.GetInvoice(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, inv)
}

func (h *InvoiceServiceHandler) ListInvoicesSvc(c *gin.Context) {
	userIDStr := c.Query("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	invoices, err := h.svc.ListInvoices(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invoices)
}
