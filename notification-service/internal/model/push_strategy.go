package model

import "time"

type PushStrategy struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Schedule      string    `json:"schedule"`
	Targeting     string    `json:"targeting"`
	FrequencyCap  int       `json:"frequency_cap"`
	DNDStartHour  int       `json:"dnd_start_hour"`
	DNDEndHour    int       `json:"dnd_end_hour"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type PushEvent struct {
	UserID    int64     `json:"user_id"`
	StrategyID int64    `json:"strategy_id"`
	Timestamp time.Time `json:"timestamp"`
}
