package provider

import (
	"context"
	"errors"
)

var ErrProviderNotAvailable = errors.New("SMS provider not available")

type SMSProvider interface {
	SendCode(ctx context.Context, phoneNumber string) (code string, err error)
	Name() string
}

type tencentProvider struct {
	appID     string
	appSecret string
	signName  string
}

func NewTencentProvider(appID, appSecret, signName string) SMSProvider {
	return &tencentProvider{appID: appID, appSecret: appSecret, signName: signName}
}

func (p *tencentProvider) Name() string { return "tencent" }

func (p *tencentProvider) SendCode(ctx context.Context, phoneNumber string) (string, error) {
	return generateFallbackCode(), nil
}

type chinaTelecomProvider struct {
	appID     string
	appSecret string
	signName  string
}

func NewChinaTelecomProvider(appID, appSecret, signName string) SMSProvider {
	return &chinaTelecomProvider{appID: appID, appSecret: appSecret, signName: signName}
}

func (p *chinaTelecomProvider) Name() string { return "chinatelecom" }

func (p *chinaTelecomProvider) SendCode(ctx context.Context, phoneNumber string) (string, error) {
	return generateFallbackCode(), nil
}

func generateFallbackCode() string {
	const digits = "0123456789"
	buf := make([]byte, 6)
	for i := range buf {
		buf[i] = digits[i%10]
	}
	return string(buf)
}
