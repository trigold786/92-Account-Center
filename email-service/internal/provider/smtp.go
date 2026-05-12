package provider

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/trigold786/92-Account-Center/email-service/internal/model"
)

type SMTPProvider struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSMTPProvider(host, port, username, password, from string) *SMTPProvider {
	return &SMTPProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (p *SMTPProvider) Name() string {
	return string(model.ProviderSMTP)
}

func (p *SMTPProvider) Send(ctx context.Context, to, subject, content string) *EmailResult {
	addr := p.host + ":" + p.port

	msg := buildMessage(p.from, to, subject, content)

	var auth smtp.Auth
	if p.username != "" && p.password != "" {
		auth = smtp.PlainAuth("", p.username, p.password, p.host)
	}

	err := smtp.SendMail(addr, auth, p.from, []string{to}, []byte(msg))
	if err != nil {
		return &EmailResult{
			Success: false,
			Error:   fmt.Errorf("smtp send failed: %w", err),
		}
	}

	return &EmailResult{
		MessageID: "smtp-" + to,
		Success:   true,
	}
}

func buildMessage(from, to, subject, body string) string {
	header := make([]byte, 0)
	header = append(header, []byte("From: "+from+"\r\n")...)
	header = append(header, []byte("To: "+to+"\r\n")...)
	header = append(header, []byte("Subject: "+subject+"\r\n")...)
	header = append(header, []byte("MIME-Version: 1.0\r\n")...)
	header = append(header, []byte("Content-Type: text/html; charset=\"utf-8\"\r\n")...)
	header = append(header, []byte("\r\n")...)
	header = append(header, []byte(body+"\r\n")...)
	return string(header)
}
