package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"device-fingerprint-service/internal/model"
	"device-fingerprint-service/internal/service"
)

// DeviceFingerprintHandler handles device fingerprint requests
type DeviceFingerprintHandler struct {
	deviceService service.DeviceFingerprintService
}

// NewDeviceFingerprintHandler creates a new DeviceFingerprintHandler
func NewDeviceFingerprintHandler(deviceService service.DeviceFingerprintService) *DeviceFingerprintHandler {
	return &DeviceFingerprintHandler{deviceService: deviceService}
}

// RegisterDevice handles device fingerprint registration
func (h *DeviceFingerprintHandler) RegisterDevice(c *gin.Context) {
	var req model.DeviceFingerprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (would be set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Convert userID to uint64 (assuming it's stored as float64 in context)
	var uid uint64
	switch v := userID.(type) {
	case float64:
		uid = uint64(v)
	case int64:
		uid = uint64(v)
	case int:
		uid = uint64(v)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	resp, err := h.deviceService.RegisterDevice(c.Request.Context(), uid, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDevice handles getting a device fingerprint
func (h *DeviceFingerprintHandler) GetDevice(c *gin.Context) {
	fingerprintID := c.Param("fingerprint_id")
	if fingerprintID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fingerprint ID is required"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var uid uint64
	switch v := userID.(type) {
	case float64:
		uid = uint64(v)
	case int64:
		uid = uint64(v)
	case int:
		uid = uint64(v)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	device, err := h.deviceService.GetDevice(c.Request.Context(), uid, fingerprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if device == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}

	c.JSON(http.StatusOK, device)
}

// ListDevices handles listing all devices for a user
func (h *DeviceFingerprintHandler) ListDevices(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var uid uint64
	switch v := userID.(type) {
	case float64:
		uid = uint64(v)
	case int64:
		uid = uint64(v)
	case int:
		uid = uint64(v)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	devices, err := h.deviceService.ListDevices(c.Request.Context(), uid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// CheckTrust handles checking if a device is trusted
func (h *DeviceFingerprintHandler) CheckTrust(c *gin.Context) {
	fingerprintID := c.Param("fingerprint_id")
	if fingerprintID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "fingerprint ID is required"})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var uid uint64
	switch v := userID.(type) {
	case float64:
		uid = uint64(v)
	case int64:
		uid = uint64(v)
	case int:
		uid = uint64(v)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	isTrusted, err := h.deviceService.IsTrusted(c.Request.Context(), uid, fingerprintID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_trusted": isTrusted})
}

// AssessRisk handles assessing risk for a device fingerprint
func (h *DeviceFingerprintHandler) AssessRisk(c *gin.Context) {
	var req model.DeviceFingerprintRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	var uid uint64
	switch v := userID.(type) {
	case float64:
		uid = uint64(v)
	case int64:
		uid = uint64(v)
	case int:
		uid = uint64(v)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
		return
	}

	isRisky, err := h.deviceService.AssessRisk(c.Request.Context(), uid, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"is_risky": isRisky})
}