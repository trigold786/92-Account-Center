package model

import "time"

type SocialAccount struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	Provider     string     `json:"provider"`
	ProviderUID  string     `json:"provider_uid"`
	Email        string     `json:"email,omitempty"`
	AvatarURL    string     `json:"avatar_url,omitempty"`
	AccessToken  string     `json:"-"`
	RefreshToken string     `json:"-"`
	ExpiresAt    *time.Time `json:"expires_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type SocialUserInfo struct {
	Provider    string `json:"provider"`
	ProviderUID string `json:"provider_uid"`
	Email       string `json:"email,omitempty"`
	AvatarURL   string `json:"avatar_url,omitempty"`
	Name        string `json:"name,omitempty"`
}
