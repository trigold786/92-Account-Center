package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/config-service/internal/model"
	"github.com/trigold786/92-Account-Center/config-service/internal/service"
)

type FAQHandler struct {
	svc *service.FAQService
}

func NewFAQHandler(svc *service.FAQService) *FAQHandler {
	return &FAQHandler{svc: svc}
}

func (h *FAQHandler) ListFAQs(c *gin.Context) {
	category := c.Query("category")
	faqs, err := h.svc.ListFAQs(c.Request.Context(), category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": faqs})
}

func (h *FAQHandler) SearchFAQs(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter q is required"})
		return
	}
	results, err := h.svc.SearchFAQs(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": results})
}

func (h *FAQHandler) CreateFAQ(c *gin.Context) {
	var req struct {
		Category  string `json:"category" binding:"required"`
		Question  string `json:"question" binding:"required"`
		Answer    string `json:"answer" binding:"required"`
		SortOrder int    `json:"sort_order"`
		Tags      string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	faq, err := h.svc.CreateFAQ(c.Request.Context(), &model.FAQ{
		Category:  req.Category,
		Question:  req.Question,
		Answer:    req.Answer,
		SortOrder: req.SortOrder,
		Tags:      req.Tags,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": faq})
}
