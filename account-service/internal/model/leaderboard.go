package model

type LeaderboardEntry struct {
	Rank        int64  `json:"rank"`
	UserID      int64  `json:"user_id"`
	Score       int64  `json:"score"`
	DisplayName string `json:"display_name"`
}

type SocialProofMessage struct {
	Message   string `json:"message"`
	Timestamp string `json:"timestamp,omitempty"`
}
