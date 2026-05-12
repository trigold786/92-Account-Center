package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/trigold786/92-Account-Center/email-service/internal/model"
)

type SESProvider struct {
	region     string
	accessKey  string
	secretKey  string
	from       string
	endpoint   string
	httpClient *http.Client
}

func NewSESProvider(region, accessKey, secretKey, from string) *SESProvider {
	return &SESProvider{
		region:    region,
		accessKey: accessKey,
		secretKey: secretKey,
		from:      from,
		endpoint:  fmt.Sprintf("https://email.%s.amazonaws.com", region),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (p *SESProvider) Name() string {
	return string(model.ProviderAWSSES)
}

type sesSendEmailInput struct {
	Source      string            `json:"Source"`
	Destination DestinationInput  `json:"Destination"`
	Message     MessageInput      `json:"Message"`
}

type DestinationInput struct {
	ToAddresses []string `json:"ToAddresses"`
}

type MessageInput struct {
	Subject SubjectInput `json:"Subject"`
	Body    BodyInput    `json:"Body"`
}

type SubjectInput struct {
	Data string `json:"Data"`
}

type BodyInput struct {
	Html BodyContent `json:"Html"`
}

type BodyContent struct {
	Data string `json:"Data"`
}

func (p *SESProvider) Send(ctx context.Context, to, subject, content string) *EmailResult {
	payload := sesSendEmailInput{
		Source: p.from,
		Destination: DestinationInput{
			ToAddresses: []string{to},
		},
		Message: MessageInput{
			Subject: SubjectInput{
				Data: subject,
			},
			Body: BodyInput{
				Html: BodyContent{
					Data: content,
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return &EmailResult{
			Success: false,
			Error:   fmt.Errorf("failed to marshal ses payload: %w", err),
		}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/v3/email/outbound-emails", strings.NewReader(string(body)))
	if err != nil {
		return &EmailResult{
			Success: false,
			Error:   fmt.Errorf("failed to create ses request: %w", err),
		}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Api-Key", p.accessKey)
	req.Header.Set("X-Secret-Key", p.secretKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return &EmailResult{
			Success: false,
			Error:   fmt.Errorf("ses request failed: %w", err),
		}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return &EmailResult{
			Success: false,
			Error:   fmt.Errorf("ses returned status %d: %s", resp.StatusCode, string(respBody)),
		}
	}

	return &EmailResult{
		MessageID: "ses-" + to,
		Success:   true,
	}
}
