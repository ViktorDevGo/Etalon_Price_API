package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
	"github.com/prokoleso/etalon-price-api/internal/logger"
	fourtochki "github.com/prokoleso/etalon-price-api/internal/providers/4tochki"
	"github.com/prokoleso/etalon-price-api/internal/providers"
	"github.com/prokoleso/etalon-price-api/internal/repository/postgres"
	"github.com/prokoleso/etalon-price-api/internal/service"
)

func main() {
	// Command line flags
	providerName := flag.String("supplier", "4tochki", "Supplier/provider name to sync")
	batchSize := flag.Int("batch-size", 0, "Batch size (0 = use config default)")
	codesStr := flag.String("codes", "", "Comma-separated list of product codes")
	codesFile := flag.String("codes-file", "", "Path to file with product codes (one per line)")
	flag.Parse()

	// Load .env file if exists
	_ = godotenv.Load()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize logger
	appLogger := logger.New(cfg.App.LogLevel)
	appLogger.Info("Starting Etalon Price synchronization service",
		"env", cfg.App.Env,
		"provider", *providerName,
	)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		appLogger.Info("Received shutdown signal, cancelling sync...")
		cancel()
	}()

	// Initialize database
	db, err := postgres.New(ctx, postgres.Config{
		DSN:    cfg.Database.DSN,
		Logger: appLogger,
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
		appLogger.Error("Failed to register 4tochki provider", "error", err)
		os.Exit(1)
	}

	appLogger.Info("Provider registered", "provider", "4tochki")

	// Initialize sync service
	syncService := service.NewSyncService(service.SyncServiceConfig{
		ProviderRegistry: registry,
		DB:               db,
		Logger:           logger.WithComponent(appLogger, "sync"),
	})

	// Parse product codes
	var codes []string

	if *codesFile != "" {
		// Read codes from file
		fileCodes, err := readCodesFromFile(*codesFile)
		if err != nil {
			appLogger.Error("Failed to read codes from file", "file", *codesFile, "error", err)
			os.Exit(1)
		}
		codes = fileCodes
		appLogger.Info("Loaded codes from file", "file", *codesFile, "count", len(codes))
	} else if *codesStr != "" {
		// Parse codes from comma-separated string
		codes = parseCodesFromString(*codesStr)
		appLogger.Info("Parsed codes from command line", "count", len(codes))
	} else {
		appLogger.Info("No specific codes provided, will sync all available products")
	}

	// Determine batch size
	syncBatchSize := cfg.Fourtochki.BatchSize
	if *batchSize > 0 {
		syncBatchSize = *batchSize
	}

	// Run synchronization
	appLogger.Info("Starting synchronization",
		"provider", *providerName,
		"batch_size", syncBatchSize,
		"codes_count", len(codes),
	)

	result, err := syncService.SyncSupplier(ctx, *providerName, service.SyncOptions{
		Codes:      codes,
		BatchSize:  syncBatchSize,
		BatchDelay: cfg.Fourtochki.BatchDelay,
	})

	if err != nil {
		appLogger.Error("Synchronization failed",
			"provider", *providerName,
			"error", err,
		)
		os.Exit(1)
	}

	appLogger.Info("Synchronization completed successfully",
		"provider", *providerName,
		"sync_run_id", result.SyncRunID,
		"total_products", result.TotalProducts,
		"success_products", result.SuccessProducts,
		"failed_products", result.FailedProducts,
		"duration", result.Duration,
		"errors_count", len(result.Errors),
	)

	if result.FailedProducts > 0 {
		appLogger.Warn("Some products failed to sync",
			"failed_count", result.FailedProducts,
		)
		os.Exit(1)
	}

	os.Exit(0)
}

// parseCodesFromString parses comma-separated codes
func parseCodesFromString(s string) []string {
	parts := strings.Split(s, ",")
	codes := make([]string, 0, len(parts))

	for _, part := range parts {
		code := strings.TrimSpace(part)
		if code != "" {
			codes = append(codes, code)
		}
	}

	return codes
}

// readCodesFromFile reads codes from file (one per line)
func readCodesFromFile(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var codes []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		code := strings.TrimSpace(scanner.Text())
		// Skip empty lines and comments
		if code != "" && !strings.HasPrefix(code, "#") {
			codes = append(codes, code)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return codes, nil
}
