package service

import (
	"crypto/tls"
	"gopkg.in/gomail.v2"
	"jian-unified-system/jus-hermes/email/config"
)

type EmailService interface {
	Send(to string, subject string, body string) error
}

type emailService struct {
	cfg config.EmailConfig
}

func DefaultEmailService() EmailService {
	return &emailService{cfg: *config.DefaultEmailConfig()}
}

func NewEmailService(cfg config.EmailConfig) EmailService {
	return &emailService{cfg: cfg}
}

func (e *emailService) Send(to string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", m.FormatAddress(e.cfg.From, e.cfg.DisplayName))
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	d := gomail.NewDialer(e.cfg.Host, e.cfg.Port, e.cfg.Username, e.cfg.Password)

	// 465 端口 SMTPS，关闭证书校验
	d.TLSConfig = &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         e.cfg.Host,
	}

	return d.DialAndSend(m)
}
