package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/compliance-service/internal/model"
	"github.com/trigold786/92-Account-Center/compliance-service/internal/service"
)

type BlacklistHandler struct {
	blacklistSvc service.BlacklistService
}

func NewBlacklistHandler(blacklistSvc service.BlacklistService) *BlacklistHandler {
	return &BlacklistHandler{blacklistSvc: blacklistSvc}
}

func (h *BlacklistHandler) AddEntry(c *gin.Context) {
	var req model.BlacklistEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	entry, err := h.blacklistSvc.AddEntry(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"code": 0, "data": entry})
}

func (h *BlacklistHandler) CheckEntry(c *gin.Context) {
	var req model.BlacklistCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	blocked, reason, err := h.blacklistSvc.CheckBlocked(c.Request.Context(), req.EntryType, req.EntryValue)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": gin.H{"blocked": blocked, "reason": reason}})
}

func (h *BlacklistHandler) RemoveEntry(c *gin.Context) {
	entryType := c.Param("type")
	entryValue := c.Param("value")
	if err := h.blacklistSvc.RemoveEntry(c.Request.Context(), entryType, entryValue); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "removed"})
}

func (h *BlacklistHandler) ListEntries(c *gin.Context) {
	entryType := c.Query("type")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	entries, err := h.blacklistSvc.ListEntries(c.Request.Context(), entryType, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": entries})
}
