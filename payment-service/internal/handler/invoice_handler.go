package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/payment-service/internal/model"
	"github.com/trigold786/92-Account-Center/payment-service/internal/repository"
)

type InvoiceHandler struct {
	repo *repository.InvoiceRepository
}

func NewInvoiceHandler(repo *repository.InvoiceRepository) *InvoiceHandler {
	return &InvoiceHandler{repo: repo}
}

func (h *InvoiceHandler) CreateInvoice(c *gin.Context) {
	var req struct {
		OrderID int64  `json:"order_id"`
		Title   string `json:"title"`
		TaxID   string `json:"tax_id"`
		Email   string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	inv := &model.Invoice{
		UserID:    userID.(int64),
		OrderID:   req.OrderID,
		InvoiceNo: "INV" + strconv.FormatInt(req.OrderID, 10),
		Title:     req.Title,
		TaxID:     req.TaxID,
		Email:     req.Email,
		Status:    "pending",
	}
	if err := h.repo.Create(c.Request.Context(), inv); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, inv)
}

func (h *InvoiceHandler) ListInvoices(c *gin.Context) {
	userID, _ := c.Get("user_id")
	invoices, err := h.repo.GetByUserID(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, invoices)
}
