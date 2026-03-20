package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
	"github.com/prokoleso/etalon-price-api/internal/email"
	"github.com/prokoleso/etalon-price-api/internal/providers/severavto"
	"github.com/prokoleso/etalon-price-api/internal/repository"
	"github.com/prokoleso/etalon-price-api/internal/service"
	"github.com/robfig/cron/v3"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Load configuration
	cfg := config.Load()

	// Setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Starting Severavto scheduler",
		"timezone", os.Getenv("TZ"),
		"base_url", cfg.SeveravtoBaseURL,
	)

	// Connect to database
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	logger.Info("Connected to database")

	// Create repositories
	nomenclRepo := repository.NewNomenclatureRepository(pool)
	pricesRepo := repository.NewPricesStockRepository(pool)
	warehouseRepo := repository.NewWarehouseRepository(pool)

	// Create email client (if enabled)
	var emailService *email.Client
	if cfg.EmailEnabled {
		emailService = email.NewClient(email.Config{
			SMTPHost: cfg.EmailSMTPHost,
			SMTPPort: cfg.EmailSMTPPort,
			Username: cfg.EmailUsername,
			Password: cfg.EmailPassword,
			From:     cfg.EmailFrom,
		}, logger)
	}

	// Create Severavto provider
	provider := severavto.NewProvider(severavto.ProviderConfig{
		BaseURL: cfg.SeveravtoBaseURL,
		APIKey:  cfg.SeveravtoAPIKey,
		Timeout: cfg.SeveravtoTimeout,
		Logger:  logger,
	})

	// Create sync service
	syncService := service.NewSeveravtoSyncService(
		provider,
		nomenclRepo,
		pricesRepo,
		warehouseRepo,
		emailService,
		cfg.EmailNotificationTo,
		logger,
	)

	// Create cron scheduler with Moscow timezone
	location, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		logger.Error("Failed to load timezone", "error", err)
		location = time.UTC
	}

	c := cron.New(cron.WithLocation(location), cron.WithLogger(
		cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags)),
	))

	// Schedule tyres sync: every 3 hours at :00
	// 0 */3 * * * = 00:00, 03:00, 06:00, 09:00, 12:00, 15:00, 18:00, 21:00 (MSK)
	_, err = c.AddFunc("0 */3 * * *", func() {
		logger.Info("Starting scheduled tyres sync")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		if err := syncService.SyncTyres(ctx); err != nil {
			logger.Error("Scheduled tyres sync failed", "error", err)
		}
	})
	if err != nil {
		log.Fatalf("Failed to schedule tyres sync: %v", err)
	}

	// Schedule rims sync: every 3 hours at :15 (15 minutes after tyres)
	// 15 */3 * * * = 00:15, 03:15, 06:15, 09:15, 12:15, 15:15, 18:15, 21:15 (MSK)
	// Это гарантирует 15-минутный перерыв между запросами шин и дисков (rate limit 10 минут)
	_, err = c.AddFunc("15 */3 * * *", func() {
		logger.Info("Starting scheduled rims sync")
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
		defer cancel()

		if err := syncService.SyncRims(ctx); err != nil {
			logger.Error("Scheduled rims sync failed", "error", err)
		}
	})
	if err != nil {
		log.Fatalf("Failed to schedule rims sync: %v", err)
	}

	// Start cron scheduler
	c.Start()
	logger.Info("Severavto scheduler started",
		"tyres_schedule", "0 */3 * * * (every 3 hours at :00)",
		"rims_schedule", "15 */3 * * * (every 3 hours at :15)",
		"timezone", location.String(),
	)

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down Severavto scheduler...")
	ctx := c.Stop()
	<-ctx.Done()
	logger.Info("Severavto scheduler stopped")
}
