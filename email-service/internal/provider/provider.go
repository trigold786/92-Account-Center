package provider

import (
	"context"
)

type EmailResult struct {
	MessageID string
	Success   bool
	Error     error
}

type EmailProvider interface {
	Send(ctx context.Context, to, subject, content string) *EmailResult
	Name() string
}
