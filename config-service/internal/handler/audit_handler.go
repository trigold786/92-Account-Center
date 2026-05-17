package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/config-service/internal/model"
	"github.com/trigold786/92-Account-Center/config-service/internal/service"
)

type AuditHandler struct {
	auditSvc service.AuditService
}

func NewAuditHandler(auditSvc service.AuditService) *AuditHandler {
	return &AuditHandler{auditSvc: auditSvc}
}

// ListLogs GET /api/v1/config/audit-logs
func (h *AuditHandler) ListLogs(c *gin.Context) {
	var filter model.AuditLogFilter
	filter.OperationType = c.Query("operation_type")
	filter.Operator = c.Query("operator")
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if startStr := c.Query("start_time"); startStr != "" {
		t, err := strconv.ParseInt(startStr, 10, 64)
		if err == nil {
			tm := time.Unix(t, 0)
			filter.StartTime = &tm
		}
	}
	if endStr := c.Query("end_time"); endStr != "" {
		t, err := strconv.ParseInt(endStr, 10, 64)
		if err == nil {
			tm := time.Unix(t, 0)
			filter.EndTime = &tm
		}
	}

	logs, total, err := h.auditSvc.ListLogs(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": logs, "total": total})
}

// GetLogByID GET /api/v1/config/audit-logs/:id
func (h *AuditHandler) GetLogByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	log, err := h.auditSvc.GetLogByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": err.Error()})
		return
	}
	if log == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3, "message": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": log})
}
