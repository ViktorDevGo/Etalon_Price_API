package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prokoleso/etalon-price-api/internal/config"
	"github.com/prokoleso/etalon-price-api/internal/email"
	"github.com/prokoleso/etalon-price-api/internal/providers/severavto"
	"github.com/prokoleso/etalon-price-api/internal/repository"
	"github.com/prokoleso/etalon-price-api/internal/service"
)

func main() {
	// Parse command-line flags
	typeFlag := flag.String("type", "all", "Sync type: tyres, rims, or all")
	flag.Parse()

	// Load configuration
	cfg := config.Load()

	// Setup structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	logger.Info("Starting Severavto synchronization",
		"type", *typeFlag,
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

	// Execute synchronization
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Minute)
	defer cancel()

	switch *typeFlag {
	case "tyres":
		logger.Info("Synchronizing tyres only")
		if err := syncService.SyncTyres(ctx); err != nil {
			logger.Error("Failed to sync tyres", "error", err)
			os.Exit(1)
		}

	case "rims":
		logger.Info("Synchronizing rims only")
		if err := syncService.SyncRims(ctx); err != nil {
			logger.Error("Failed to sync rims", "error", err)
			os.Exit(1)
		}

	case "all":
		logger.Info("Synchronizing both tyres and rims")

		// Sync tyres
		if err := syncService.SyncTyres(ctx); err != nil {
			logger.Error("Failed to sync tyres", "error", err)
			os.Exit(1)
		}

		// Wait 15 minutes before syncing rims (rate limiting: 10 минут между запросами)
		logger.Info("Waiting 15 minutes before syncing rims (rate limit)")
		time.Sleep(15 * time.Minute)

		// Sync rims
		if err := syncService.SyncRims(ctx); err != nil {
			logger.Error("Failed to sync rims", "error", err)
			os.Exit(1)
		}

	default:
		fmt.Printf("Unknown sync type: %s. Use 'tyres', 'rims', or 'all'\n", *typeFlag)
		os.Exit(1)
	}

	logger.Info("Severavto synchronization completed successfully")
}
