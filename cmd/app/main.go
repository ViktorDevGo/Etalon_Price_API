package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
	"github.com/prokoleso/etalon-price-api/internal/httpserver"
	"github.com/prokoleso/etalon-price-api/internal/logger"
	fourtochki "github.com/prokoleso/etalon-price-api/internal/providers/4tochki"
	"github.com/prokoleso/etalon-price-api/internal/providers"
	"github.com/prokoleso/etalon-price-api/internal/repository/postgres"
	"github.com/prokoleso/etalon-price-api/internal/service"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger := logger.New(cfg.App.LogLevel)
	appLogger.Info("Starting Etalon Price API server",
		"env", cfg.App.Env,
		"port", cfg.HTTP.Port,
	)

	// Create context
	ctx := context.Background()

	// Initialize database
	db, err := postgres.New(ctx, postgres.Config{
		DSN:    cfg.Database.DSN,
		Logger: logger.WithComponent(appLogger, "database"),
	})
	if err != nil {
		appLogger.Error("Failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	appLogger.Info("Database connection established")

	// Initialize provider registry
	registry := providers.NewRegistry()

	// Register 4tochki provider
	fourtochkiProvider := fourtochki.NewProvider(fourtochki.ProviderConfig{
		WSDLURL:    cfg.Fourtochki.WSDLURL,
		Login:      cfg.Fourtochki.Login,
		Password:   cfg.Fourtochki.Password,
		Timeout:    cfg.Fourtochki.Timeout,
		RetryCount: cfg.Fourtochki.RetryCount,
		RetryDelay: cfg.Fourtochki.RetryDelay,
		Logger:     logger.WithProvider(appLogger, "4tochki"),
	})

	if err := registry.Register(fourtochkiProvider); err != nil {
		appLogger.Error("Failed to register provider", "error", err)
		os.Exit(1)
	}

	appLogger.Info("Provider registered", "provider", "4tochki")

	// Initialize sync service
	syncService := service.NewSyncService(service.SyncServiceConfig{
		ProviderRegistry: registry,
		DB:               db,
		Logger:           logger.WithComponent(appLogger, "sync"),
	})

	// Initialize HTTP server
	server := httpserver.New(httpserver.Config{
		Port:        cfg.HTTP.Port,
		SyncService: syncService,
		DB:          db,
		Registry:    registry,
		Logger:      logger.WithComponent(appLogger, "http"),
	})

	// Start server in goroutine
	go func() {
		appLogger.Info("HTTP server starting", "port", cfg.HTTP.Port)
		if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("HTTP server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	appLogger.Info("Shutting down HTTP server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		appLogger.Error("HTTP server shutdown error", "error", err)
		os.Exit(1)
	}

	appLogger.Info("HTTP server stopped gracefully")
	os.Exit(0)
}
