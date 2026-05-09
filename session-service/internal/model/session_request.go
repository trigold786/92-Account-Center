package model

type CreateSessionRequest struct {
	UserID           int64  `json:"user_id" binding:"required"`
	DeviceFingerprint string `json:"device_fingerprint" binding:"required"`
	IPAddress        string `json:"ip_address" binding:"required"`
}

type ValidateSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

type GetUserSessionsRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

type SessionResponse struct {
	SessionID        string `json:"session_id"`
	UserID           int64  `json:"user_id"`
	DeviceFingerprint string `json:"device_fingerprint"`
	IPAddress        string `json:"ip_address"`
	CreatedAt        string `json:"created_at"`
	LastAccessedAt   string `json:"last_accessed_at"`
	ExpiresAt        string `json:"expires_at"`
	RemainingTTL     int64  `json:"remaining_ttl"`
	IsValid          bool   `json:"is_valid"`
}

type InvalidateSessionRequest struct {
	SessionID string `json:"session_id"`
	UserID    int64  `json:"user_id"`
}

type InvalidateAllUserSessionsRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

type RefreshSessionRequest struct {
	SessionID string `json:"session_id" binding:"required"`
}

type CountUserSessionsRequest struct {
	UserID int64 `json:"user_id" binding:"required"`
}

type CountUserSessionsResponse struct {
	Count int64 `json:"count"`
}
