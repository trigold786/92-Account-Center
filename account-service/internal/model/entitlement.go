package model

type Entitlement struct {
	ID          int64   `json:"id"`
	UserID      int64   `json:"user_id"`
	FeatureCode string  `json:"feature_code"`
	TotalQuota  int     `json:"total_quota"`
	UsedQuota   int     `json:"used_quota"`
	ResetTime   *string `json:"reset_time,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type EntitlementQuota struct {
	Total int   `json:"total"`
	Used  int   `json:"used"`
	Reset int64 `json:"reset,omitempty"`
}

type ConsumeRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	FeatureCode string `json:"feature_code" binding:"required"`
	Amount      int    `json:"amount" binding:"required,min=1"`
}

type ConsumeResponse struct {
	Success   bool `json:"success"`
	Remaining int  `json:"remaining"`
}

type GrantRequest struct {
	UserID    string `json:"user_id" binding:"required"`
	TierLevel int    `json:"tier_level" binding:"required,oneof=2 3 4"`
}
