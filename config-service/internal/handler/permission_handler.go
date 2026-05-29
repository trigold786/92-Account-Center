package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/config-service/internal/model"
	"github.com/trigold786/92-Account-Center/config-service/internal/service"
)

type PermissionHandler struct {
	permSvc service.PermissionService
}

func NewPermissionHandler(permSvc service.PermissionService) *PermissionHandler {
	return &PermissionHandler{permSvc: permSvc}
}

// ListRoles GET /api/v1/config/roles
func (h *PermissionHandler) ListRoles(c *gin.Context) {
	roles, err := h.permSvc.ListRoles(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": roles})
}

// CreateRole POST /api/v1/config/roles
func (h *PermissionHandler) CreateRole(c *gin.Context) {
	var role model.Role
	if err := c.ShouldBindJSON(&role); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid request body"})
		return
	}
	operator := c.GetString("operator")
	if err := h.permSvc.CreateRole(c.Request.Context(), &role, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": role})
}

// GetRolePermissions GET /api/v1/config/roles/:id/permissions
func (h *PermissionHandler) GetRolePermissions(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	perms, err := h.permSvc.GetRolePermissions(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": perms})
}

// AddRolePermission POST /api/v1/config/roles/:id/permissions
func (h *PermissionHandler) AddRolePermission(c *gin.Context) {
	roleID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	var rp model.RolePermission
	if err := c.ShouldBindJSON(&rp); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid request body"})
		return
	}
	rp.RoleID = roleID
	operator := c.GetString("operator")
	if err := h.permSvc.AddRolePermission(c.Request.Context(), &rp, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": rp})
}

// GetUserRoles GET /api/v1/config/users/:userId/roles
func (h *PermissionHandler) GetUserRoles(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid user id"})
		return
	}
	urs, err := h.permSvc.GetUserRoles(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": urs})
}

// RemoveRole DELETE /api/v1/config/roles/:id
func (h *PermissionHandler) RemoveRole(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	operator := c.GetString("operator")
	if err := h.permSvc.DeleteRole(c.Request.Context(), id, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// RemoveRolePermission DELETE /api/v1/config/roles/:id/permissions/:permId
func (h *PermissionHandler) RemoveRolePermission(c *gin.Context) {
	permID, err := strconv.ParseInt(c.Param("permId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid permission id"})
		return
	}
	operator := c.GetString("operator")
	if err := h.permSvc.RemoveRolePermission(c.Request.Context(), permID, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// RemoveUserRole DELETE /api/v1/config/users/:userId/roles/:roleId
func (h *PermissionHandler) RemoveUserRole(c *gin.Context) {
	userID := c.Param("userId")
	roleID, err := strconv.ParseInt(c.Param("roleId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid role id"})
		return
	}
	operator := c.GetString("operator")
	if err := h.permSvc.RemoveUserRole(c.Request.Context(), userID, roleID, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// GetUserPermissions GET /api/v1/config/users/:userId/permissions
func (h *PermissionHandler) GetUserPermissions(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid user id"})
		return
	}

	operator := c.GetString("operator")
	if operator != userID {
		allowed, err := h.permSvc.CheckPermission(c.Request.Context(), operator, "permission.manage")
		if err != nil || !allowed {
			c.JSON(http.StatusForbidden, gin.H{"code": 5, "message": "permission denied"})
			return
		}
	}

	perms, err := h.permSvc.GetUserPermissions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": perms})
}

// SetUserRole POST /api/v1/config/users/:userId/roles
func (h *PermissionHandler) SetUserRole(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid user id"})
		return
	}
	var ur model.UserRole
	if err := c.ShouldBindJSON(&ur); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid request body"})
		return
	}
	ur.UserID = userID
	operator := c.GetString("operator")
	if err := h.permSvc.SetUserRole(c.Request.Context(), &ur, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": ur})
}
