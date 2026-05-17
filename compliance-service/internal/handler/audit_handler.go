package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/service"
)

const (
	DefaultRetentionDays = 180
)

type AuditHandler struct {
	auditService service.AuditService
}

func NewAuditHandler(auditService service.AuditService) *AuditHandler {
	return &AuditHandler{auditService: auditService}
}

func (h *AuditHandler) RecordLog(c *gin.Context) {
	var req model.AuditLogEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request body"})
		return
	}

	log, err := h.auditService.RecordLog(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to record log"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    201,
		"message": "Audit log recorded successfully",
		"data": model.AuditLogResponse{
			LogID:   log.LogID,
			SM3Hash: log.SM3Hash,
		},
	})
}

func (h *AuditHandler) RecordBatch(c *gin.Context) {
	var req model.BatchAuditLogRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request body"})
		return
	}

	response, err := h.auditService.RecordBatch(c.Request.Context(), req.Entries)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to record batch logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "Batch audit logs recorded",
		"data":    response,
	})
}

func (h *AuditHandler) GetLogsByUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "user_id is required"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.auditService.GetLogsByUser(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to get logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"logs":   logs,
			"limit":  limit,
			"offset": offset,
		},
	})
}

func (h *AuditHandler) GetLogsByTimeRange(c *gin.Context) {
	startStr := c.Query("start_time")
	endStr := c.Query("end_time")

	if startStr == "" || endStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "start_time and end_time are required"})
		return
	}

	startTime, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid start_time format, use RFC3339"})
		return
	}

	endTime, err := time.Parse(time.RFC3339, endStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid end_time format, use RFC3339"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	logs, err := h.auditService.GetLogsByTimeRange(c.Request.Context(), startTime, endTime, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to get logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data": gin.H{
			"logs":       logs,
			"start_time": startTime,
			"end_time":   endTime,
			"limit":      limit,
			"offset":     offset,
		},
	})
}

func (h *AuditHandler) VerifyLogIntegrity(c *gin.Context) {
	logID := c.Param("log_id")
	if logID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "log_id is required"})
		return
	}

	result, err := h.auditService.VerifyLogIntegrity(c.Request.Context(), logID)
	if err != nil {
		if err == service.ErrLogNotFound {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "log not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to verify log"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "success",
		"data":    result,
	})
}

func (h *AuditHandler) CleanupOldLogs(c *gin.Context) {
	retentionDays := DefaultRetentionDays
	if retentionStr := c.Query("retention_days"); retentionStr != "" {
		parsed, err := strconv.Atoi(retentionStr)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid retention_days"})
			return
		}
		retentionDays = parsed
	}

	result, err := h.auditService.CleanupOldLogs(c.Request.Context(), retentionDays)
	if err != nil {
		if err == service.ErrInvalidRetention {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid retention days"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to cleanup logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "cleanup completed",
		"data":    result,
	})
}
