package model

type RFMScore struct {
	UserID             int64   `json:"user_id"`
	RecencyScore       int     `json:"recency_score"`
	FrequencyScore     int     `json:"frequency_score"`
	MonetaryScore      int     `json:"monetary_score"`
	RFMSegment         string  `json:"rfm_segment"`
	RFMSegmentCN       string  `json:"rfm_segment_cn"`
	LastSubscriptionAt string  `json:"last_subscription_at"`
	TotalSubscriptions int     `json:"total_subscriptions"`
	TotalSpent         float64 `json:"total_spent"`
}

type RFMBatchRequest struct {
	UserIDs []int64 `json:"user_ids" binding:"required"`
}

type DashboardOverview struct {
	TotalUsers           int                `json:"total_users"`
	TotalSubscriptions   int                `json:"total_subscriptions"`
	TotalCreditsEarned   float64            `json:"total_credits_earned"`
	TotalCreditsConsumed float64            `json:"total_credits_consumed"`
	BlacklistActive      int                `json:"blacklist_entries_active"`
	RegistrationTrend    []DailyCount       `json:"registration_trend"`
	CreditFlow           map[string]float64 `json:"credit_flow"`
	RFMDistribution      map[string]int     `json:"rfm_distribution"`
}

type DailyCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

type FunnelStep struct {
	Name       string  `json:"name"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

type SubscriptionFunnel struct {
	Steps []FunnelStep `json:"steps"`
}

type SubscriptionStats struct {
	Freq      int
	Monetary  float64
	LastSubAt string
}

type UserTierCount struct {
	Tier  int
	Count int
}
