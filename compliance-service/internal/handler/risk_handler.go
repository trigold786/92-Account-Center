package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/service"
)

type RiskHandler struct {
	riskService *service.RiskService
}

func NewRiskHandler(riskService *service.RiskService) *RiskHandler {
	return &RiskHandler{riskService: riskService}
}

func (h *RiskHandler) RegisterRoutes(r *gin.Engine) {
	risk := r.Group("/api/v1/risk")
	{
		risk.POST("/assess", h.AssessRisk)
		risk.GET("/history/:user_id", h.GetRiskHistory)
		risk.GET("/event/:event_id", h.GetRiskEvent)
	}
}

func (h *RiskHandler) AssessRisk(c *gin.Context) {
	var req model.AssessRiskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request body"})
		return
	}

	resp, err := h.riskService.AssessRisk(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": resp})
}

func (h *RiskHandler) GetRiskHistory(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "user_id is required"})
		return
	}

	startStr := c.Query("start_date")
	endStr := c.Query("end_date")
	limitStr := c.Query("limit")

	var start, end time.Time
	var limit int = 100

	if startStr != "" {
		parsed, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			start = parsed
		}
	}

	if endStr != "" {
		parsed, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			end = parsed
		}
	}

	if limitStr != "" {
		if l, err := parseInt(limitStr); err == nil {
			limit = l
		}
	}

	if start.IsZero() {
		start = time.Now().AddDate(0, -1, 0)
	}
	if end.IsZero() {
		end = time.Now()
	}

	events, err := h.riskService.GetRiskHistory(c.Request.Context(), userID, start, end)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "internal error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": gin.H{"events": events, "limit": limit}})
}

func (h *RiskHandler) GetRiskEvent(c *gin.Context) {
	eventID := c.Param("event_id")
	if eventID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "event_id is required"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "not implemented", "data": gin.H{"event_id": eventID}})
}

func parseInt(s string) (int, error) {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, nil
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
