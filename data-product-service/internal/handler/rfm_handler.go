package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/data-product-service/internal/model"
	"github.com/trigold786/92-Account-Center/data-product-service/internal/service"
)

type RFMHandler struct {
	rfmSvc service.RFMService
}

func NewRFMHandler(rfmSvc service.RFMService) *RFMHandler {
	return &RFMHandler{rfmSvc: rfmSvc}
}

func (h *RFMHandler) GetRFM(c *gin.Context) {
	userIDStr := c.Param("user_id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid user_id"})
		return
	}

	rfm, err := h.rfmSvc.GetRFM(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": rfm})
}

func (h *RFMHandler) GetRFMBatch(c *gin.Context) {
	var req model.RFMBatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	results, err := h.rfmSvc.GetRFMBatch(c.Request.Context(), req.UserIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "success", "data": results})
}
