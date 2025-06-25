package utils

import (
	"context"
	"fmt"
	"net/smtp"

	"github.com/histopathai/auth-service/config" // Import the config package
)

// MailService deines the interface for sending emails.
type EmailService interface {
	SendEmail(ctx context.Context, recipientEmail, subject, body string) error
}

// Removed the duplicate SMTPConfig struct here.
// We will now use config.SMTPConfig directly.

type MailServiceImpl struct {
	config config.SMTPConfig // Use config.SMTPConfig
}

// NewMailService now accepts config.SMTPConfig
func NewMailService(cfg config.SMTPConfig) *MailServiceImpl {
	return &MailServiceImpl{
		config: cfg,
	}
}

// SendEmail sends an email using the SMTP configuration.
func (s *MailServiceImpl) SendEmail(ctx context.Context, recipientEmail, subject, body string) error {
	// E-posta başlıkları ve gövdesi
	msg := []byte("To: " + recipientEmail + "\r\n" +
		"From: " + s.config.Sender + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\r\n" + // HTML içerik için
		"\r\n" +
		body)

	addr := s.config.Host + ":" + fmt.Sprintf("%d", s.config.Port) // Port is int, convert to string
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	err := smtp.SendMail(addr, auth, s.config.Sender, []string{recipientEmail}, msg)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}
