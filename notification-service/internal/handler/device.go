package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
	"github.com/trigold786/92-Account-Center/notification-service/internal/service"
)

type DeviceHandler struct {
	pushService *service.PushService
}

func NewDeviceHandler(pushService *service.PushService) *DeviceHandler {
	return &DeviceHandler{pushService: pushService}
}

func (h *DeviceHandler) RegisterDevice(c *gin.Context) {
	var req model.DeviceTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}

	if err := h.pushService.RegisterDeviceToken(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "设备注册失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "设备注册成功"})
}

func (h *DeviceHandler) UnregisterDevice(c *gin.Context) {
	userID := c.Param("user_id")
	deviceToken := c.Param("device_token")

	if userID == "" || deviceToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}

	if err := h.pushService.UnregisterDeviceToken(c.Request.Context(), userID, deviceToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "设备注销失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "设备注销成功"})
}

func (h *DeviceHandler) SendPush(c *gin.Context) {
	var req model.PushSendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请求参数错误"})
		return
	}

	resp, err := h.pushService.SendPushNotification(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "推送发送失败"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
