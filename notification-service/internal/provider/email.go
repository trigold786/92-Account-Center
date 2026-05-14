package provider

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
)

type VerificationEmailProvider interface {
	SendVerificationCode(ctx context.Context, to, code string) error
	Name() string
}

type simpleSMTPProvider struct {
	host     string
	port     string
	username string
	password string
	from     string
}

func NewSimpleSMTPProvider(host, port, username, password, from string) VerificationEmailProvider {
	return &simpleSMTPProvider{
		host:     host,
		port:     port,
		username: username,
		password: password,
		from:     from,
	}
}

func (p *simpleSMTPProvider) Name() string { return "smtp-verification" }

func (p *simpleSMTPProvider) SendVerificationCode(ctx context.Context, to, code string) error {
	addr := p.host + ":" + p.port
	subject := "验证码 - 账户中心"
	body := fmt.Sprintf(
		"您的验证码是：%s\n\n验证码有效期为5分钟，请勿将验证码告知他人。\n\n如非本人操作，请忽略此邮件。",
		code,
	)

	msg := strings.Join([]string{
		"From: " + p.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	auth := smtp.PlainAuth("", p.username, p.password, p.host)

	tlsConfig := &tls.Config{
		ServerName: p.host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("smtp: tls dial %s: %w", addr, err)
	}

	client, err := smtp.NewClient(conn, p.host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp: create client: %w", err)
	}

	if err := client.Auth(auth); err != nil {
		client.Close()
		return fmt.Errorf("smtp: auth: %w", err)
	}

	if err := client.Mail(p.from); err != nil {
		client.Close()
		return fmt.Errorf("smtp: mail from: %w", err)
	}

	if err := client.Rcpt(to); err != nil {
		client.Close()
		return fmt.Errorf("smtp: rcpt to: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		client.Close()
		return fmt.Errorf("smtp: data: %w", err)
	}

	if _, err := w.Write([]byte(msg)); err != nil {
		w.Close()
		client.Close()
		return fmt.Errorf("smtp: write body: %w", err)
	}

	if err := w.Close(); err != nil {
		client.Close()
		return fmt.Errorf("smtp: close body: %w", err)
	}

	client.Quit()
	return nil
}
