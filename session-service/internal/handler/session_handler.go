package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/sunxi/92-Account-Center/session-service/internal/model"
	"github.com/sunxi/92-Account-Center/session-service/internal/service"
)

type SessionHandler struct {
	sessionService service.SessionService
}

func NewSessionHandler(sessionService service.SessionService) *SessionHandler {
	return &SessionHandler{sessionService: sessionService}
}

func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req model.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := h.sessionService.CreateSession(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"session_id": session.SessionID,
		"user_id":    session.UserID,
		"expires_at": session.ExpiresAt,
	})
}

func (h *SessionHandler) ValidateSession(c *gin.Context) {
	var req model.ValidateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionInfo, err := h.sessionService.ValidateSession(c.Request.Context(), req.SessionID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrSessionNotFound || err == service.ErrSessionExpired {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":    sessionInfo.SessionID,
		"user_id":       sessionInfo.UserID,
		"is_valid":      true,
		"remaining_ttl": sessionInfo.RemainingTTL,
	})
}

func (h *SessionHandler) GetUserSessions(c *gin.Context) {
	userIDStr := c.Param("user_id")
	var userID int64
	if _, err := parseUserID(userIDStr, &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	sessions, err := h.sessionService.GetUserSessions(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var response []gin.H
	for _, session := range sessions {
		response = append(response, gin.H{
			"session_id":         session.SessionID,
			"user_id":            session.UserID,
			"device_fingerprint": session.DeviceFingerprint,
			"ip_address":        session.IPAddress,
			"created_at":        session.CreatedAt,
			"last_accessed_at":   session.LastAccessedAt,
			"expires_at":         session.ExpiresAt,
			"remaining_ttl":     session.RemainingTTL,
		})
	}

	c.JSON(http.StatusOK, gin.H{"sessions": response})
}

func (h *SessionHandler) InvalidateSession(c *gin.Context) {
	var req model.InvalidateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.SessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	err := h.sessionService.InvalidateSession(c.Request.Context(), req.SessionID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrSessionNotFound {
			status = http.StatusNotFound
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session invalidated"})
}

func (h *SessionHandler) InvalidateAllUserSessions(c *gin.Context) {
	var req model.InvalidateAllUserSessionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.sessionService.InvalidateAllUserSessions(c.Request.Context(), req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "all sessions invalidated"})
}

func (h *SessionHandler) RefreshSession(c *gin.Context) {
	var req model.RefreshSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sessionInfo, err := h.sessionService.RefreshSession(c.Request.Context(), req.SessionID)
	if err != nil {
		status := http.StatusInternalServerError
		if err == service.ErrSessionNotFound || err == service.ErrSessionExpired {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"session_id":    sessionInfo.SessionID,
		"user_id":       sessionInfo.UserID,
		"expires_at":    sessionInfo.ExpiresAt,
		"remaining_ttl": sessionInfo.RemainingTTL,
	})
}

func parseUserID(s string, userID *int64) (bool, error) {
	var id int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return false, nil
		}
		id = id*10 + int64(c-'0')
	}
	*userID = id
	return true, nil
}
