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
