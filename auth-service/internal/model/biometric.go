
package model

import "time"

type BiometricType string

const (
	BiometricTypeFingerprint BiometricType = "fingerprint"
	BiometricTypeFace        BiometricType = "face"
	BiometricTypeIris        BiometricType = "iris"
)

type BiometricRegisterRequest struct {
	UserID        string        `json:"user_id" binding:"required"`
	DeviceFingerprint string    `json:"device_fingerprint" binding:"required"`
	BiometricType BiometricType `json:"biometric_type" binding:"required,oneof=fingerprint face iris"`
	BiometricToken string       `json:"biometric_token" binding:"required"`
}

type BiometricLoginRequest struct {
	BiometricToken string `json:"biometric_token" binding:"required"`
	DeviceFingerprint string `json:"device_fingerprint" binding:"required"`
}

type BiometricCredential struct {
	ID               string        `json:"id"`
	UserID           string        `json:"user_id"`
	DeviceFingerprint string      `json:"device_fingerprint"`
	BiometricType    BiometricType `json:"biometric_type"`
	BiometricTokenHash string    `json:"-"`
	IsActive         bool          `json:"is_active"`
	CreatedAt        time.Time     `json:"created_at"`
	LastUsedAt       *time.Time    `json:"last_used_at"`
}

