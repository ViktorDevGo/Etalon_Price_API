package main

import (
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
	"github.com/prokoleso/etalon-price-api/internal/email"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Check if email is enabled
	if !cfg.Email.Enabled {
		log.Fatal("Email is disabled in config (EMAIL_ENABLED=false)")
	}

	// Create email client
	emailClient := email.NewClient(email.Config{
		SMTPHost: cfg.Email.SMTPHost,
		SMTPPort: cfg.Email.SMTPPort,
		Username: cfg.Email.Username,
		Password: cfg.Email.Password,
		From:     cfg.Email.From,
	}, logger)

	// Build test message
	result := email.NomenclatureSyncResult{
		Type:            "tyres",
		NewCount:        150,
		SkippedCount:    53819,
		TotalInDB:       53969,
		Duration:        45 * time.Second,
		StartedAt:       time.Now().Add(-45 * time.Second),
		CompletedAt:     time.Now(),
		FilteredOutType: "Грузовая, Спецшина",
		FilteredCount:   6679,
	}

	msg := email.BuildNomenclatureSyncHTMLMessage(result)

	// Override recipient if configured
	if cfg.Email.NotificationTo != "" {
		msg.To = []string{cfg.Email.NotificationTo}
	}

	log.Printf("Sending test email to: %v", msg.To)
	log.Printf("Subject: %s", msg.Subject)
	log.Printf("SMTP: %s:%d", cfg.Email.SMTPHost, cfg.Email.SMTPPort)
	log.Printf("From: %s", cfg.Email.From)

	// Send email
	if err := emailClient.Send(msg); err != nil {
		log.Fatalf("Failed to send email: %v", err)
	}

	log.Println("✅ Test email sent successfully!")
}
