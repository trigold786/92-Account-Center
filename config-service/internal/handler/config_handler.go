package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/config-service/internal/model"
	"github.com/trigold786/92-Account-Center/config-service/internal/service"
)

type ConfigHandler struct {
	configSvc service.ConfigService
}

func NewConfigHandler(configSvc service.ConfigService) *ConfigHandler {
	return &ConfigHandler{configSvc: configSvc}
}

// ListGroups GET /api/v1/config/groups
func (h *ConfigHandler) ListGroups(c *gin.Context) {
	groups, err := h.configSvc.ListGroups(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": groups})
}

// GetGroupByID GET /api/v1/config/groups/:id
func (h *ConfigHandler) GetGroupByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	group, err := h.configSvc.GetGroupByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	if group == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3, "message": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": group})
}

// CreateGroup POST /api/v1/config/groups
func (h *ConfigHandler) CreateGroup(c *gin.Context) {
	var group model.ConfigGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid request body"})
		return
	}
	operator := c.GetString("operator")
	if err := h.configSvc.CreateGroup(c.Request.Context(), &group, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": group})
}

// UpdateGroup PUT /api/v1/config/groups/:id
func (h *ConfigHandler) UpdateGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	var group model.ConfigGroup
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid request body"})
		return
	}
	group.ID = id
	operator := c.GetString("operator")
	if err := h.configSvc.UpdateGroup(c.Request.Context(), &group, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": group})
}

// DeleteGroup DELETE /api/v1/config/groups/:id
func (h *ConfigHandler) DeleteGroup(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	operator := c.GetString("operator")
	if err := h.configSvc.DeleteGroup(c.Request.Context(), id, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// ListItems GET /api/v1/config/items
func (h *ConfigHandler) ListItems(c *gin.Context) {
	var filter model.ConfigItemFilter
	filter.GroupID = parseInt64Ptr(c.Query("group_id"))
	filter.Code = c.Query("code")
	filter.Name = c.Query("name")
	filter.DataType = c.Query("data_type")
	filter.Page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	filter.PageSize, _ = strconv.Atoi(c.DefaultQuery("page_size", "20"))

	items, total, err := h.configSvc.ListItems(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": items, "total": total})
}

// GetItemByID GET /api/v1/config/items/:id
func (h *ConfigHandler) GetItemByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	item, err := h.configSvc.GetItemByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3, "message": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": item})
}

// CreateItem POST /api/v1/config/items
func (h *ConfigHandler) CreateItem(c *gin.Context) {
	var item model.ConfigItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid request body"})
		return
	}
	operator := c.GetString("operator")
	if err := h.configSvc.CreateItem(c.Request.Context(), &item, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": item})
}

// UpdateItem PUT /api/v1/config/items/:id
func (h *ConfigHandler) UpdateItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	var item model.ConfigItem
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid request body"})
		return
	}
	item.ID = id
	operator := c.GetString("operator")
	changeReason := c.Query("change_reason")
	if err := h.configSvc.UpdateItem(c.Request.Context(), &item, changeReason, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": item})
}

// DeleteItem DELETE /api/v1/config/items/:id
func (h *ConfigHandler) DeleteItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	operator := c.GetString("operator")
	if err := h.configSvc.DeleteItem(c.Request.Context(), id, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// ResetItemToDefault POST /api/v1/config/items/:id/reset-default
func (h *ConfigHandler) ResetItemToDefault(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	operator := c.GetString("operator")
	if err := h.configSvc.ResetItemToDefault(c.Request.Context(), id, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// ListVersions GET /api/v1/config/items/:id/versions
func (h *ConfigHandler) ListVersions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	versions, err := h.configSvc.ListVersionsByItemID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": versions})
}

// GetItemByCode GET /api/v1/config/items/by-code/:code (internal)
func (h *ConfigHandler) GetItemByCode(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid code"})
		return
	}
	item, err := h.configSvc.GetItemByCode(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	if item == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3, "message": "not found"})
		return
	}
	if item.IsSensitive {
		item.CurrentValue = "***"
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": item})
}

func parseInt64Ptr(s string) *int64 {
	if s == "" {
		return nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &v
}
