package model

import "time"

type Template struct {
	ID        int64     `json:"id"`
	Channel   string    `json:"channel"`
	Name      string    `json:"name"`
	Subject   string    `json:"subject,omitempty"`
	Body      string    `json:"body"`
	Variables string    `json:"variables,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SendRecord struct {
	ID         int64     `json:"id"`
	TemplateID int64     `json:"template_id"`
	Channel    string    `json:"channel"`
	Recipient  string    `json:"recipient"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}
