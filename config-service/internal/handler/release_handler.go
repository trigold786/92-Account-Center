package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/config-service/internal/model"
	"github.com/trigold786/92-Account-Center/config-service/internal/service"
)

type ReleaseHandler struct {
	releaseSvc service.ReleaseService
}

func NewReleaseHandler(releaseSvc service.ReleaseService) *ReleaseHandler {
	return &ReleaseHandler{releaseSvc: releaseSvc}
}

// ListReleases GET /api/v1/config/releases
func (h *ReleaseHandler) ListReleases(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	releases, total, err := h.releaseSvc.ListReleases(c.Request.Context(), status, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": releases, "total": total})
}

// GetReleaseByID GET /api/v1/config/releases/:id
func (h *ReleaseHandler) GetReleaseByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	rel, err := h.releaseSvc.GetReleaseByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	if rel == nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 3, "message": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": rel})
}

// CreateRelease POST /api/v1/config/releases
func (h *ReleaseHandler) CreateRelease(c *gin.Context) {
	var rel model.ConfigRelease
	if err := c.ShouldBindJSON(&rel); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid request body"})
		return
	}
	operator := c.GetString("operator")
	if err := h.releaseSvc.CreateRelease(c.Request.Context(), &rel, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": rel})
}

// SubmitRelease PUT /api/v1/config/releases/:id/submit
func (h *ReleaseHandler) SubmitRelease(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	operator := c.GetString("operator")
	if err := h.releaseSvc.SubmitRelease(c.Request.Context(), id, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// ApproveRelease PUT /api/v1/config/releases/:id/approve
func (h *ReleaseHandler) ApproveRelease(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	operator := c.GetString("operator")
	if err := h.releaseSvc.ApproveRelease(c.Request.Context(), id, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// RejectRelease PUT /api/v1/config/releases/:id/reject
func (h *ReleaseHandler) RejectRelease(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	operator := c.GetString("operator")
	if err := h.releaseSvc.RejectRelease(c.Request.Context(), id, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// ExecuteRelease POST /api/v1/config/releases/:id/execute
func (h *ReleaseHandler) ExecuteRelease(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	operator := c.GetString("operator")
	if err := h.releaseSvc.ExecuteRelease(c.Request.Context(), id, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok"})
}

// ListReleaseItems GET /api/v1/config/releases/:id/items
func (h *ReleaseHandler) ListReleaseItems(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	items, err := h.releaseSvc.ListReleaseItems(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": items})
}

// AddReleaseItem POST /api/v1/config/releases/:id/items
func (h *ReleaseHandler) AddReleaseItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid id"})
		return
	}
	var ri model.ConfigReleaseItem
	if err := c.ShouldBindJSON(&ri); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "invalid request body"})
		return
	}
	ri.ReleaseID = id
	operator := c.GetString("operator")
	if err := h.releaseSvc.AddReleaseItem(c.Request.Context(), &ri, operator); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 2, "message": "internal error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": "ok", "data": ri})
}
