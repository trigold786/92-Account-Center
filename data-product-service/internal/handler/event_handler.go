package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type EventHandler struct {
	svc *service.EventService
}

func NewEventHandler(svc *service.EventService) *EventHandler {
	return &EventHandler{svc: svc}
}

func (h *EventHandler) TrackEvent(c *gin.Context) {
	var req struct {
		EventType  string                 `json:"event_type"`
		SessionID  string                 `json:"session_id"`
		DeviceID   string                 `json:"device_id"`
		Properties map[string]interface{} `json:"properties"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	event := model.Event{
		EventType:  req.EventType,
		UserID:     userID.(int64),
		SessionID:  req.SessionID,
		DeviceID:   req.DeviceID,
		Properties: req.Properties,
	}
	if err := h.svc.ValidateEvent(c.Request.Context(), event.EventType, event.Properties); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.BatchProcess(c.Request.Context(), []model.Event{event}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "tracked"})
}

func (h *EventHandler) BatchTrack(c *gin.Context) {
	var req struct {
		Events []struct {
			EventType  string                 `json:"event_type"`
			SessionID  string                 `json:"session_id"`
			DeviceID   string                 `json:"device_id"`
			Properties map[string]interface{} `json:"properties"`
		} `json:"events"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID, _ := c.Get("user_id")
	var events []model.Event
	for _, e := range req.Events {
		events = append(events, model.Event{
			EventType:  e.EventType,
			UserID:     userID.(int64),
			SessionID:  e.SessionID,
			DeviceID:   e.DeviceID,
			Properties: e.Properties,
		})
	}
	if err := h.svc.BatchProcess(c.Request.Context(), events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "tracked", "count": len(events)})
}
