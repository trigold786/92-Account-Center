package model

import "time"

type FAQ struct {
	ID         int64     `json:"id"`
	Category   string    `json:"category"`
	Question   string    `json:"question"`
	Answer     string    `json:"answer"`
	SortOrder  int       `json:"sort_order"`
	Tags       string    `json:"tags"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
