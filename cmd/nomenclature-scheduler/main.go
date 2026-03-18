package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
	"github.com/prokoleso/etalon-price-api/internal/repository"
	"github.com/prokoleso/etalon-price-api/internal/service"
	"github.com/robfig/cron/v3"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg := config.Load()

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Connect to database
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create services
	nomenclatureRepo := repository.NewNomenclatureRepository(pool)
	nomenclatureService := service.NewNomenclatureService(nomenclatureRepo, logger)

	// Create cron scheduler
	c := cron.New(cron.WithLogger(cron.VerbosePrintfLogger(log.New(os.Stdout, "cron: ", log.LstdFlags))))

	// Schedule nomenclature sync daily at 2 AM
	_, err = c.AddFunc("0 2 * * *", func() {
		logger.Info("Starting scheduled nomenclature sync")
		ctx := context.Background()

		// Sync tyres
		if err := nomenclatureService.SyncTyres(ctx); err != nil {
			logger.Error("Failed to sync tyres", "error", err)
		}

		// Sync rims
		if err := nomenclatureService.SyncRims(ctx); err != nil {
			logger.Error("Failed to sync rims", "error", err)
		}

		logger.Info("Scheduled nomenclature sync completed")
	})
	if err != nil {
		log.Fatalf("Failed to schedule nomenclature job: %v", err)
	}

	// Start scheduler
	c.Start()
	logger.Info("Scheduler started", "nomenclature", "daily at 2 AM")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	logger.Info("Shutting down scheduler...")

	// Stop scheduler gracefully
	ctx := c.Stop()
	<-ctx.Done()

	logger.Info("Scheduler stopped")
}
