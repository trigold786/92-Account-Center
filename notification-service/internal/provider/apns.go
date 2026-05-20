package provider

import (
	"context"
	"fmt"
	"log"
	"time"
)

type APNsConfig struct {
	CertificatePath string
	KeyPath         string
	BundleID        string
	Production      bool
}

type apnsProvider struct {
	config APNsConfig
}

func NewAPNsProvider(cfg APNsConfig) PushProvider {
	return &apnsProvider{config: cfg}
}

func (p *apnsProvider) Name() string { return "apns" }

func (p *apnsProvider) Send(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	log.Printf("[APNs] sending to device %s: title=%s body=%s", req.DeviceToken, req.Title, req.Body)
	msgID := fmt.Sprintf("apns-%d", time.Now().UnixNano())
	return &PushResponse{
		MessageID: msgID,
		Success:   true,
	}, nil
}

func (p *apnsProvider) ValidateToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("apns: device token is empty")
	}
	return nil
}
