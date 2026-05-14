package model

import (
	"time"
)

type EmailProvider string

const (
	ProviderSendGrid EmailProvider = "sendgrid"
	ProviderAWSSES   EmailProvider = "aws_ses"
	ProviderSMTP     EmailProvider = "smtp"
)

type EmailStatus string

const (
	EmailStatusSent   EmailStatus = "sent"
	EmailStatusFailed EmailStatus = "failed"
)

type EmailLog struct {
	EmailLogID string       `json:"email_log_id" db:"email_log_id"`
	UserID     *int64       `json:"user_id" db:"user_id"`
	Email      string       `json:"email" db:"email"`
	Subject    string       `json:"subject" db:"subject"`
	Content    string       `json:"content" db:"content"`
	Status     EmailStatus  `json:"status" db:"status"`
	Provider   EmailProvider `json:"provider" db:"provider"`
	CreatedAt  time.Time    `json:"created_at" db:"created_at"`
}

func (EmailLog) TableName() string {
	return "email_logs"
}
