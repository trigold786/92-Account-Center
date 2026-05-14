package model

import "time"

type LoginRequest struct {
	Credential            string `json:"credential" binding:"required"`
	Password              string `json:"password,omitempty"`
	Code                  string `json:"code,omitempty"`
	MagicLink             string `json:"magic_link,omitempty"`
	DeviceFingerprintID   string `json:"device_fingerprint_id,omitempty"`
}

type LoginResponse struct {
	AccessToken       string `json:"access_token"`
	RefreshToken      string `json:"refresh_token"`
	ExpiresIn         int64  `json:"expires_in"`
	UserID            int64  `json:"user_id"`
	AccountID         string `json:"account_id"`
	IsTrusted         bool   `json:"is_trusted"`
	IsTrustedDevice   bool   `json:"is_trusted_device"`
	MFARequired       bool   `json:"mfa_required"`
	TokenID           string `json:"token_id,omitempty"`
	DeviceBindingInfo string `json:"device_binding_info,omitempty"`
}

type User struct {
	ID               int64      `json:"id" db:"id"`
	PhoneNumber      string     `json:"phone_number" db:"phone_number"`
	AccountID        string     `json:"account_id" db:"account_id"`
	Email            string     `json:"email,omitempty" db:"email"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	MFAEnabled       bool       `json:"mfa_enabled" db:"mfa_enabled"`
	MFASecret        string     `json:"-" db:"mfa_secret"`
	LastStrongAuthAt *time.Time `json:"last_strong_auth_at,omitempty" db:"last_strong_auth_at"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
