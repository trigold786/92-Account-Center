package provider

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"

	"github.com/trigold786/92-Account-Center/notification-service/internal/model"
)

type SESProvider struct {
	client *sesv2.Client
	from   string
}

func NewSESProvider(region, accessKey, secretKey, from string) (*SESProvider, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return &SESProvider{
		client: sesv2.NewFromConfig(cfg),
		from:   from,
	}, nil
}

func (p *SESProvider) Name() string {
	return string(model.ProviderAWSSES)
}

func (p *SESProvider) Send(ctx context.Context, to, subject, content string) *EmailResult {
	_, err := p.client.SendEmail(ctx, &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(p.from),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{Data: aws.String(subject)},
				Body: &types.Body{
					Html: &types.Content{Data: aws.String(content)},
				},
			},
		},
	})
	if err != nil {
		return &EmailResult{
			Success: false,
			Error:   fmt.Errorf("ses send email: %w", err),
		}
	}
	return &EmailResult{
		MessageID: "ses-" + to,
		Success:   true,
	}
}
