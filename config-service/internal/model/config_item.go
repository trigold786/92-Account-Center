package model

import "time"

type ConfigItem struct {
	ID            int64     `json:"id"`
	GroupID       int64     `json:"group_id"`
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	DataType      string    `json:"data_type"`
	CurrentValue  string    `json:"current_value"`
	DefaultValue  string    `json:"default_value"`
	MinValue      string    `json:"min_value"`
	MaxValue      string    `json:"max_value"`
	AllowedValues string    `json:"allowed_values"`
	IsSensitive   bool      `json:"is_sensitive"`
	IsEnabled     bool      `json:"is_enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type ConfigItemFilter struct {
	GroupID   *int64
	Code      string
	Name      string
	DataType  string
	IsEnabled *bool
	Page      int
	PageSize  int
}
