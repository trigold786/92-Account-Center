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
