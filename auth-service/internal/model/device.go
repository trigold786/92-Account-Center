package model

type DeviceFingerprint struct {
	ID            uint64  `json:"id"`
	UserID        uint64  `json:"user_id"`
	FingerprintID string  `json:"fingerprint_id"`
	UserAgent     string  `json:"user_agent"`
	IPAddress     string  `json:"ip_address"`
	Country       string  `json:"country"`
	City          string  `json:"city"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Features      []byte  `json:"features"`
	IsTrusted     bool    `json:"is_trusted"`
	LastUsedAt    int64   `json:"last_used_at"`
	CreatedAt     int64   `json:"created_at"`
	UpdatedAt     int64   `json:"updated_at"`
}

type DeviceFingerprintRequest struct {
	FingerprintID string  `json:"fingerprint_id" binding:"required"`
	UserAgent     string  `json:"user_agent"`
	IPAddress     string  `json:"ip_address"`
	Country       string  `json:"country"`
	City          string  `json:"city"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Features      []byte  `json:"features"`
}

type DeviceFingerprintResponse struct {
	ID            uint64 `json:"id"`
	FingerprintID string `json:"fingerprint_id"`
	IsTrusted     bool   `json:"is_trusted"`
	LastUsedAt    int64  `json:"last_used_at"`
}
