package notifications

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

type SMTPConfig struct {
	Host string
	Port int
	User string
	Pass string
	From string
}

// SendEmail sends a plain-text alert email via SMTP.
func SendEmail(ctx context.Context, cfg SMTPConfig, to, subject, body string) error {
	if cfg.Host == "" {
		return fmt.Errorf("SMTP host not configured")
	}

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	auth := smtp.PlainAuth("", cfg.User, cfg.Pass, cfg.Host)

	from := stripCRLF(cfg.From)
	to = stripCRLF(to)
	subject = stripCRLF(subject)

	msg := []byte(
		"From: " + from + "\r\n" +
			"To: " + to + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"Content-Type: text/plain; charset=UTF-8\r\n" +
			"\r\n" +
			body + "\r\n",
	)

	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

// stripCRLF removes carriage-return and newline characters so values placed
// in SMTP headers can't be used to inject extra headers or recipients.
func stripCRLF(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return s
}
