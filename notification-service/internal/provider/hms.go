package provider

import (
	"context"
	"fmt"
	"log"
	"time"
)

type HMSConfig struct {
	AppID     string
	AppSecret string
}

type hmsProvider struct {
	config HMSConfig
}

func NewHMSProvider(cfg HMSConfig) PushProvider {
	return &hmsProvider{config: cfg}
}

func (p *hmsProvider) Name() string { return "hms" }

func (p *hmsProvider) Send(ctx context.Context, req *PushRequest) (*PushResponse, error) {
	log.Printf("[HMS] sending to device %s: title=%s body=%s", req.DeviceToken, req.Title, req.Body)
	msgID := fmt.Sprintf("hms-%d", time.Now().UnixNano())
	return &PushResponse{
		MessageID: msgID,
		Success:   true,
	}, nil
}

func (p *hmsProvider) ValidateToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("hms: device token is empty")
	}
	return nil
}
