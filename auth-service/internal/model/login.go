package model

// LoginRequest represents a login request
type LoginRequest struct {
	Credential      string `json:"credential" binding:"required"`
	Password        string `json:"password,omitempty"`
	Code            string `json:"code,omitempty"` // For SMS/email OTP
	MagicLink       string `json:"magic_link,omitempty"` // For magic link login
	DeviceFingerprintID string `json:"device_fingerprint_id,omitempty"` // Device fingerprint for trust checking
}

// LoginResponse represents a login response
type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	UserID       int64  `json:"user_id"`
	AccountID    string `json:"account_id"`
}

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}