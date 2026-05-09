package provider

import (
	"context"
	"fmt"

	"github.com/sunxi/92-Account-Center/email-service/internal/model"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridProvider struct {
	apiKey string
	from   string
}

func NewSendGridProvider(apiKey, from string) *SendGridProvider {
	return &SendGridProvider{
		apiKey: apiKey,
		from:   from,
	}
}

func (p *SendGridProvider) Name() string {
	return string(model.ProviderSendGrid)
}

func (p *SendGridProvider) Send(ctx context.Context, to, subject, content string) *EmailResult {
	from := mail.NewEmail("Account Center", p.from)
	toEmail := mail.NewEmail("", to)
	message := mail.NewSingleEmail(from, subject, toEmail, content, content)

	client := sendgrid.NewSendClient(p.apiKey)
	response, err := client.Send(message)
	if err != nil {
		return &EmailResult{
			Success: false,
			Error:   fmt.Errorf("sendgrid send failed: %w", err),
		}
	}

	if response.StatusCode >= 400 {
		return &EmailResult{
			Success: false,
			Error:   fmt.Errorf("sendgrid returned status %d: %s", response.StatusCode, response.Body),
		}
	}

	return &EmailResult{
		MessageID: response.Headers["X-Message-Id"],
		Success:   true,
	}
}
