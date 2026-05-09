package model

// DeviceFingerprint represents a device fingerprint record
type DeviceFingerprint struct {
	ID             uint64    `json:"id"`
	UserID         uint64    `json:"user_id"`
	FingerprintID  string    `json:"fingerprint_id"` // Unique identifier from FingerprintJS
	UserAgent      string    `json:"user_agent"`
	IPAddress      string    `json:"ip_address"`
	Country        string    `json:"country"`
	City           string    `json:"city"`
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Features       []byte    `json:"features"` // Encrypted or hashed device features
	IsTrusted      bool      `json:"is_trusted"`
	LastUsedAt     int64     `json:"last_used_at"` // Unix timestamp
	CreatedAt      int64     `json:"created_at"`
	UpdatedAt      int64     `json:"updated_at"`
}

// DeviceFingerprintRequest represents the request to register/update a device fingerprint
type DeviceFingerprintRequest struct {
	FingerprintID string `json:"fingerprint_id" binding:"required"`
	UserAgent     string `json:"user_agent"`
	IPAddress     string `json:"ip_address"`
	Country       string `json:"city"` // Note: keeping as per spec, though usually country and city are separate
	City          string `json:"city"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	Features      []byte  `json:"features"`
}

// DeviceFingerprintResponse represents the response for device fingerprint operations
type DeviceFingerprintResponse struct {
	ID          uint64 `json:"id"`
	FingerprintID string `json:"fingerprint_id"`
	IsTrusted   bool   `json:"is_trusted"`
	LastUsedAt  int64  `json:"last_used_at"`
}