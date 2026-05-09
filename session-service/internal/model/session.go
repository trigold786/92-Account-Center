package model

import "time"

type Session struct {
	SessionID        string    `json:"session_id"`
	UserID           int64     `json:"user_id"`
	DeviceFingerprint string   `json:"device_fingerprint"`
	IPAddress        string    `json:"ip_address"`
	CreatedAt        time.Time `json:"created_at"`
	LastAccessedAt   time.Time `json:"last_accessed_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	IsActive         bool      `json:"is_active"`
}

type SessionInfo struct {
	SessionID        string    `json:"session_id"`
	UserID           int64     `json:"user_id"`
	DeviceFingerprint string   `json:"device_fingerprint"`
	IPAddress        string    `json:"ip_address"`
	CreatedAt        time.Time `json:"created_at"`
	LastAccessedAt   time.Time `json:"last_accessed_at"`
	ExpiresAt        time.Time `json:"expires_at"`
	RemainingTTL     int64     `json:"remaining_ttl"`
}

func (s *Session) ToSessionInfo(remainingTTL int64) *SessionInfo {
	return &SessionInfo{
		SessionID:        s.SessionID,
		UserID:           s.UserID,
		DeviceFingerprint: s.DeviceFingerprint,
		IPAddress:        s.IPAddress,
		CreatedAt:        s.CreatedAt,
		LastAccessedAt:   s.LastAccessedAt,
		ExpiresAt:        s.ExpiresAt,
		RemainingTTL:     remainingTTL,
	}
}
