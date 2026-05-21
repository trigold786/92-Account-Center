package model

import "time"

type Plan struct {
	ID          int64                  `json:"id"`
	Name        string                 `json:"name"`
	DisplayName string                 `json:"display_name"`
	Price       float64                `json:"price"`
	Interval    string                 `json:"interval"`
	Features    map[string]interface{} `json:"features"`
	Active      bool                   `json:"active"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}
