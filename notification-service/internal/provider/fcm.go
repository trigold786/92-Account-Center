package provider

import (
	"context"
	"fmt"
	"log"
	"time"
)

type FCMConfig struct {
	ServerKey string
	ProjectID string
}

type fcmProvider struct {
	config FCMConfig
}

func NewFCMProvider(cfg FCMConfig) PushProvider {
	return &fcmProvider{config: cfg}
}

func (p *fcmProvider) Name() string { return "fcm" }

func (p *fcmProvider) Send(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	log.Printf("[FCM] sending to device %s: title=%s body=%s", req.DeviceToken, req.Title, req.Body)
	msgID := fmt.Sprintf("fcm-%d", time.Now().UnixNano())
	return &PushResponse{
		MessageID: msgID,
		Success:   true,
	}, nil
}

func (p *fcmProvider) ValidateToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("fcm: device token is empty")
	}
	return nil
}
