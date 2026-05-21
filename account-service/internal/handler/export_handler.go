package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type ExportHandler struct {
	svc *service.ExportService
}

func NewExportHandler(svc *service.ExportService) *ExportHandler {
	return &ExportHandler{svc: svc}
}

func (h *ExportHandler) ExportPersonalData(c *gin.Context) {
	userID, _ := c.Get("user_id")
	export, err := h.svc.ExportPersonalData(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=personal_data.json")
	c.JSON(http.StatusOK, export)
}

func (h *ExportHandler) RequestExport(c *gin.Context) {
	userID, _ := c.Get("user_id")
	reqID, err := h.svc.RequestExport(c.Request.Context(), userID.(int64))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, gin.H{"request_id": reqID, "status": "processing"})
}
