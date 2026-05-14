package provider

import (
	"context"
	"fmt"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
)

type SendGridProvider struct {
	apiKey string
	from   string
}

func NewSendGridProvider(apiKey, from string) *SendGridProvider {
	return &SendGridProvider{apiKey: apiKey, from: from}
}

func (p *SendGridProvider) Name() string {
	return string(model.ProviderSendGrid)
}

func (p *SendGridProvider) Send(ctx context.Context, to, subject, content string) *EmailResult {
	return &EmailResult{
		Success: false,
		Error:   fmt.Errorf("sendgrid provider not available: install github.com/sendgrid/sendgrid-go"),
	}
}
