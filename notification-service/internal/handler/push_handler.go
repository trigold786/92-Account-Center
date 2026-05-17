package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
	"github.com/trigold786/92-Account-Center/notification-service/internal/service"
)

type PushHandler struct {
	pushService *service.PushService
}

func NewPushHandler(pushService *service.PushService) *PushHandler {
	return &PushHandler{pushService: pushService}
}

func (h *PushHandler) SendPush(c *gin.Context) {
	var req model.PushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.PushResponse{
			Code:    400,
			Message: "请求参数错误",
		})
		return
	}

	resp, err := h.pushService.SendPush(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.PushResponse{
			Code:    500,
			Message: "推送发送失败",
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *PushHandler) RegisterDevice(c *gin.Context) {
	var req model.PushDevice
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.PushResponse{
			Code:    400,
			Message: "请求参数错误",
		})
		return
	}

	if err := h.pushService.RegisterDevice(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusInternalServerError, model.PushResponse{
			Code:    500,
			Message: "设备注册失败",
		})
		return
	}

	c.JSON(http.StatusOK, model.PushResponse{
		Code:    200,
		Message: "设备注册成功",
	})
}

func (h *PushHandler) GetUserDevices(c *gin.Context) {
	userID := c.Param("user_id")

	devices, err := h.pushService.GetUserDevices(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.PushResponse{
			Code:    500,
			Message: "获取设备列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, model.PushResponse{
		Code:    200,
		Message: "获取成功",
		Data:    gin.H{"devices": devices},
	})
}
