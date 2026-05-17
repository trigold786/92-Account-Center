package model

import "time"

type ConfigVersion struct {
	ID           int64     `json:"id"`
	ItemID       int64     `json:"item_id"`
	ValueBefore  string    `json:"value_before"`
	ValueAfter   string    `json:"value_after"`
	ChangeReason string    `json:"change_reason"`
	ChangedBy    string    `json:"changed_by"`
	CreatedAt    time.Time `json:"created_at"`
}
