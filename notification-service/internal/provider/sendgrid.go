package provider

import (
	"context"
	"fmt"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"

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
	from := mail.NewEmail("", p.from)
	toAddr := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toAddr, content, content)
	client := sendgrid.NewSendClient(p.apiKey)
	response, err := client.Send(message)
	if err != nil {
		return &EmailResult{
			Success: false,
			Error:   fmt.Errorf("sendgrid send: %w", err),
		}
	}
	if response.StatusCode >= 400 {
		return &EmailResult{
			Success: false,
			Error:   fmt.Errorf("sendgrid error: status=%d body=%s", response.StatusCode, response.Body),
		}
	}
	return &EmailResult{
		MessageID: "sendgrid-" + to,
		Success:   true,
	}
}
