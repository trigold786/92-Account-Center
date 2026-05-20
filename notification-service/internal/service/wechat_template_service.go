package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
)

type WeChatTemplateService struct {
	client interface{}
	logger *slog.Logger
}

var validTemplates = map[string]bool{
	"subscription_expiring": true,
	"payment_success":       true,
	"referral_bonus":        true,
	"tier_upgrade":          true,
}

func NewWeChatTemplateService(client interface{}) *WeChatTemplateService {
	return &WeChatTemplateService{client: client, logger: slog.Default()}
}

func (s *WeChatTemplateService) ValidateTemplateID(templateType string) error {
	if !validTemplates[templateType] {
		return errors.New("invalid template type: " + templateType)
	}
	return nil
}

func (s *WeChatTemplateService) SendSubscriptionMessage(ctx context.Context, openID, templateType string, data map[string]string) error {
	if err := s.ValidateTemplateID(templateType); err != nil {
		return err
	}
	s.logger.Info("sending WeChat subscribe message", "openID", openID, "type", templateType, "data", data)
	return nil
}

func (s *WeChatTemplateService) SendWithRetry(ctx context.Context, openID, templateType string, data map[string]string) error {
	var err error
	for i := 0; i < 3; i++ {
		err = s.SendSubscriptionMessage(ctx, openID, templateType, data)
		if err == nil {
			return nil
		}
		s.logger.Warn("retry sending template message", "attempt", i+1, "error", err)
	}
	return fmt.Errorf("failed after 3 retries: %w", err)
}
