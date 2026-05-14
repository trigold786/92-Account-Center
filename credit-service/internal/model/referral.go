package model

type ReferralRelation struct {
	ID                       int64  `json:"id"`
	ReferrerID               int64  `json:"referrer_id"`
	RefereeID                int64  `json:"referee_id"`
	RefereeSubscriptionCount int    `json:"referee_subscription_count"`
	Status                   string `json:"status"`
	CreatedAt                string `json:"created_at"`
	UpdatedAt                string `json:"updated_at"`
}

type ReferralBindRequest struct {
	ReferrerCode string `json:"referrer_code" binding:"required"`
	RefereeID    string `json:"referee_id" binding:"required"`
}

type GenerateLinkResponse struct {
	ReferralCode string `json:"referral_code"`
	ReferralLink string `json:"referral_link"`
}

type ReferralSummary struct {
	TotalReferees  int     `json:"total_referees"`
	TotalEarned    float64 `json:"total_earned"`
	ActiveReferees int     `json:"active_referees"`
}
