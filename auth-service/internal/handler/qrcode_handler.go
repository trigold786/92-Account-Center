package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/trigold786/92-Account-Center/auth-service/internal/model"
	"github.com/trigold786/92-Account-Center/auth-service/internal/service"
)

type QRCodeHandler struct {
	qrcodeService *service.QRCodeService
}

func NewQRCodeHandler(qrcodeService *service.QRCodeService) *QRCodeHandler {
	return &QRCodeHandler{qrcodeService: qrcodeService}
}

func (h *QRCodeHandler) Generate(c *gin.Context) {
	resp, err := h.qrcodeService.Generate(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *QRCodeHandler) GetStatus(c *gin.Context) {
	codeID := c.Param("code_id")
	if codeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code_id is required"})
		return
	}

	resp, err := h.qrcodeService.GetStatus(c.Request.Context(), codeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *QRCodeHandler) Scan(c *gin.Context) {
	codeID := c.Param("code_id")
	if codeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code_id is required"})
		return
	}

	var req model.QRCodeScanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.CodeID = codeID

	userID, err := strconv.ParseInt(c.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		userID = req.UserID
	}

	if err := h.qrcodeService.Scan(c.Request.Context(), codeID, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "QR code scanned successfully"})
}

func (h *QRCodeHandler) Confirm(c *gin.Context) {
	codeID := c.Param("code_id")
	if codeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "code_id is required"})
		return
	}

	var req model.QRCodeConfirmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.CodeID = codeID

	userID, err := strconv.ParseInt(c.GetHeader("X-User-ID"), 10, 64)
	if err != nil {
		userID = req.UserID
	}

	resp, err := h.qrcodeService.Confirm(c.Request.Context(), codeID, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
