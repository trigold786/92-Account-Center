package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
	"github.com/trigold786/92-Account-Center/notification-service/internal/service"
)

type TemplateHandler struct {
	svc *service.TemplateService
}

func NewTemplateHandler(svc *service.TemplateService) *TemplateHandler {
	return &TemplateHandler{svc: svc}
}

func (h *TemplateHandler) CreateTemplate(c *gin.Context) {
	var req struct {
		Channel   string `json:"channel" binding:"required"`
		Name      string `json:"name" binding:"required"`
		Subject   string `json:"subject"`
		Body      string `json:"body" binding:"required"`
		Variables string `json:"variables"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	tpl, err := h.svc.CreateTemplate(c.Request.Context(), &model.Template{
		Channel:   req.Channel,
		Name:      req.Name,
		Subject:   req.Subject,
		Body:      req.Body,
		Variables: req.Variables,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, tpl)
}

func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	tpl, err := h.svc.GetTemplate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tpl)
}

func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	channel := c.Query("channel")
	templates, err := h.svc.ListTemplates(c.Request.Context(), channel)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *TemplateHandler) UpdateTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var req struct {
		Channel   string `json:"channel" binding:"required"`
		Name      string `json:"name" binding:"required"`
		Subject   string `json:"subject"`
		Body      string `json:"body" binding:"required"`
		Variables string `json:"variables"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	tpl, err := h.svc.UpdateTemplate(c.Request.Context(), &model.Template{
		ID:        id,
		Channel:   req.Channel,
		Name:      req.Name,
		Subject:   req.Subject,
		Body:      req.Body,
		Variables: req.Variables,
	})
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tpl)
}

func (h *TemplateHandler) DeleteTemplate(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.svc.DeleteTemplate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

func (h *TemplateHandler) ListSendRecords(c *gin.Context) {
	templateID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid template id"})
		return
	}

	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	records, err := h.svc.ListSendRecords(c.Request.Context(), templateID, offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"records": records})
}
