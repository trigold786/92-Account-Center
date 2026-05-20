package model

import "time"

type PushPlatform string

const (
	PushPlatformIOS     PushPlatform = "ios"
	PushPlatformAndroid PushPlatform = "android"
	PushPlatformWeb     PushPlatform = "web"
	PushPlatformXiaomi  PushPlatform = "xiaomi"
	PushPlatformHuawei  PushPlatform = "huawei"
	PushPlatformOppo    PushPlatform = "oppo"
	PushPlatformVivo    PushPlatform = "vivo"
)

type PushType string

const (
	PushTypeSecurity PushType = "security"
	PushTypeLogin    PushType = "login"
	PushTypeVerify   PushType = "verify"
	PushTypeSystem   PushType = "system"
)

type PushRequest struct {
	UserID      string                 `json:"user_id" binding:"required"`
	Platform    PushPlatform           `json:"platform" binding:"required,oneof=ios android web xiaomi huawei oppo vivo"`
	Type        PushType               `json:"type" binding:"required,oneof=security login verify system"`
	Title       string                 `json:"title" binding:"required"`
	Body        string                 `json:"body" binding:"required"`
	Data        map[string]interface{} `json:"data"`
	DeviceToken string                 `json:"device_token"`
}

type PushResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PushDevice struct {
	ID           string       `json:"id"`
	UserID       string       `json:"user_id"`
	DeviceToken  string       `json:"device_token"`
	Platform     PushPlatform `json:"platform"`
	DeviceName   string       `json:"device_name"`
	IsActive     bool         `json:"is_active"`
	LastActiveAt *time.Time   `json:"last_active_at"`
	CreatedAt    *time.Time   `json:"created_at"`
}

type DeviceTokenRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	DeviceToken string `json:"device_token" binding:"required"`
	Platform    string `json:"platform" binding:"required,oneof=ios android huawei"`
}

type PushSendRequest struct {
	UserID   string            `json:"user_id" binding:"required"`
	Title    string            `json:"title" binding:"required"`
	Body     string            `json:"body" binding:"required"`
	Data     map[string]string `json:"data,omitempty"`
	Platform string            `json:"platform,omitempty"`
}

type DeviceToken struct {
	UserID       string `json:"user_id"`
	DeviceToken  string `json:"device_token"`
	Platform     string `json:"platform"`
	RegisteredAt string `json:"registered_at"`
}
