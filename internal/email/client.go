package email

import (
	"crypto/tls"
	"fmt"
	"log/slog"

	"gopkg.in/gomail.v2"
)

// Client handles email sending
type Client struct {
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
	logger   *slog.Logger
}

// Config holds email configuration
type Config struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
	From     string
}

// NewClient creates a new email client
func NewClient(cfg Config, logger *slog.Logger) *Client {
	return &Client{
		smtpHost: cfg.SMTPHost,
		smtpPort: cfg.SMTPPort,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
		logger:   logger,
	}
}

// Message represents an email message
type Message struct {
	To      []string
	Subject string
	Body    string
	HTML    bool
}

// Send sends an email message
func (c *Client) Send(msg Message) error {
	c.logger.Info("Sending email",
		"to", msg.To,
		"subject", msg.Subject,
	)

	m := gomail.NewMessage()
	m.SetHeader("From", c.from)
	m.SetHeader("To", msg.To...)
	m.SetHeader("Subject", msg.Subject)

	if msg.HTML {
		m.SetBody("text/html", msg.Body)
	} else {
		m.SetBody("text/plain", msg.Body)
	}

	d := gomail.NewDialer(c.smtpHost, c.smtpPort, c.username, c.password)

	// Configure TLS
	d.TLSConfig = &tls.Config{
		ServerName:         c.smtpHost,
		InsecureSkipVerify: false,
	}

	// For port 465, use implicit SSL/TLS
	if c.smtpPort == 465 {
		d.SSL = true
	}

	// Try to send the email
	if err := d.DialAndSend(m); err != nil {
		c.logger.Error("Failed to send email",
			"error", err,
			"to", msg.To,
		)
		return fmt.Errorf("failed to send email: %w", err)
	}

	c.logger.Info("Email sent successfully",
		"to", msg.To,
		"subject", msg.Subject,
	)

	return nil
}
