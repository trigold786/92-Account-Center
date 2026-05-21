package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/trigold786/92-Account-Center/account-service/internal/service"
)

type OpenAPIHandler struct {
	svc *service.OpenAPIService
}

func NewOpenAPIHandler(svc *service.OpenAPIService) *OpenAPIHandler {
	return &OpenAPIHandler{svc: svc}
}

func (h *OpenAPIHandler) IssueToken(c *gin.Context) {
	var req struct {
		ClientID string `json:"client_id"`
		Scope    string `json:"scope"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.svc.IssueToken(c.Request.Context(), req.ClientID, req.Scope)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, token)
}
