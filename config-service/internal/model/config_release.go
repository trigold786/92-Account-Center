package model

import "time"

type ConfigRelease struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedBy   string     `json:"created_by"`
	ApprovedBy  string     `json:"approved_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ApprovedAt  *time.Time `json:"approved_at"`
	ReleasedAt  *time.Time `json:"released_at"`
}

type ConfigReleaseItem struct {
	ID          int64     `json:"id"`
	ReleaseID   int64     `json:"release_id"`
	ItemID      int64     `json:"item_id"`
	ValueBefore string    `json:"value_before"`
	ValueAfter  string    `json:"value_after"`
	ChangeReason string   `json:"change_reason"`
	CreatedAt   time.Time `json:"created_at"`
}
