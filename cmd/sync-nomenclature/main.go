package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
	"github.com/prokoleso/etalon-price-api/internal/email"
	"github.com/prokoleso/etalon-price-api/internal/repository"
	"github.com/prokoleso/etalon-price-api/internal/service"
)

func main() {
	// Parse flags
	syncType := flag.String("type", "all", "Sync type: 'tyres', 'rims', or 'all'")
	flag.Parse()

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

	// Create email client
	var emailClient *email.Client
	if cfg.EmailEnabled {
		emailClient = email.NewClient(email.Config{
			SMTPHost: cfg.EmailSMTPHost,
			SMTPPort: cfg.EmailSMTPPort,
			Username: cfg.EmailUsername,
			Password: cfg.EmailPassword,
			From:     cfg.EmailFrom,
		}, logger)
		logger.Info("Email notifications enabled", "to", cfg.EmailNotificationTo)
	} else {
		logger.Info("Email notifications disabled")
	}

	// Create repository and service
	nomenclatureRepo := repository.NewNomenclatureRepository(pool)
	nomenclatureService := service.NewNomenclatureService(nomenclatureRepo, logger)

	ctx := context.Background()

	fmt.Println("================================================================")
	fmt.Println("Синхронизация номенклатуры 4tochki")
	fmt.Println("================================================================")
	fmt.Println()

	// Execute sync based on type
	switch *syncType {
	case "tyres":
		logger.Info("Starting tyres sync")
		result, err := nomenclatureService.SyncTyresWithResult(ctx)
		if err != nil {
			log.Fatalf("Failed to sync tyres: %v", err)
		}
		printStats(result)
		sendEmailNotification(emailClient, cfg, result, logger)

	case "rims":
		logger.Info("Starting rims sync")
		result, err := nomenclatureService.SyncRimsWithResult(ctx)
		if err != nil {
			log.Fatalf("Failed to sync rims: %v", err)
		}
		printStats(result)
		sendEmailNotification(emailClient, cfg, result, logger)

	case "all":
		logger.Info("Starting full sync")

		// Sync tyres
		tyresResult, err := nomenclatureService.SyncTyresWithResult(ctx)
		if err != nil {
			log.Fatalf("Failed to sync tyres: %v", err)
		}
		printStats(tyresResult)
		sendEmailNotification(emailClient, cfg, tyresResult, logger)

		// Sync rims
		rimsResult, err := nomenclatureService.SyncRimsWithResult(ctx)
		if err != nil {
			log.Fatalf("Failed to sync rims: %v", err)
		}
		printStats(rimsResult)
		sendEmailNotification(emailClient, cfg, rimsResult, logger)

	default:
		log.Fatalf("Invalid sync type: %s (use 'tyres', 'rims', or 'all')", *syncType)
	}

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println("Синхронизация завершена успешно")
	fmt.Println("================================================================")
}

func printStats(result *service.SyncResult) {
	fmt.Println()
	fmt.Println("Статистика:")
	fmt.Println("-----------------------------------------------------------------")

	var typeRu string
	if result.Type == "tyres" {
		typeRu = "шин"
	} else {
		typeRu = "дисков"
	}

	fmt.Printf("Всего %s в БД: %d\n", typeRu, result.TotalInDB)
	fmt.Printf("Добавлено новых: %d\n", result.NewCount)
	fmt.Printf("Пропущено дублей: %d\n", result.SkippedCount)

	if result.FilteredCount > 0 {
		fmt.Printf("Отфильтровано (%s): %d\n", result.FilteredOutType, result.FilteredCount)
	}
}

func sendEmailNotification(emailClient *email.Client, cfg *config.Config, result *service.SyncResult, logger *slog.Logger) {
	if emailClient == nil || !cfg.EmailEnabled {
		return
	}

	// Build email message
	emailResult := email.NomenclatureSyncResult{
		Type:            result.Type,
		NewCount:        result.NewCount,
		SkippedCount:    result.SkippedCount,
		TotalInDB:       result.TotalInDB,
		Duration:        result.Duration,
		StartedAt:       result.StartedAt,
		CompletedAt:     result.CompletedAt,
		FilteredOutType: result.FilteredOutType,
		FilteredCount:   result.FilteredCount,
	}

	msg := email.BuildNomenclatureSyncHTMLMessage(emailResult)

	// Override recipient if configured
	if cfg.EmailNotificationTo != "" {
		msg.To = []string{cfg.EmailNotificationTo}
	}

	// Send email
	if err := emailClient.Send(msg); err != nil {
		logger.Error("Failed to send email notification", "error", err)
	}
}
