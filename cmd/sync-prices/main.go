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
	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
	fourtochki "github.com/prokoleso/etalon-price-api/internal/providers/4tochki"
	"github.com/prokoleso/etalon-price-api/internal/repository"
	"github.com/prokoleso/etalon-price-api/internal/service"
)

func main() {
	// Parse flags
	syncType := flag.String("type", "all", "Sync type: 'tyres', 'rims', or 'all'")
	flag.Parse()

	// Load .env file if exists
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Setup logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	// Connect to database
	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Create 4tochki API client
	fourtochkiClient := fourtochki.NewClient(fourtochki.ClientConfig{
		WSDLURL:    cfg.Fourtochki.WSDLURL,
		Login:      cfg.Fourtochki.Login,
		Password:   cfg.Fourtochki.Password,
		Timeout:    cfg.Fourtochki.Timeout,
		RetryCount: cfg.Fourtochki.RetryCount,
		RetryDelay: cfg.Fourtochki.RetryDelay,
		Logger:     logger,
	})

	// Create repositories and service
	pricesStockRepo := repository.NewPricesStockRepository(pool)
	warehouseRepo := repository.NewWarehouseRepository(pool)
	pricesStockService := service.NewPricesStockService(pricesStockRepo, warehouseRepo, fourtochkiClient, logger)

	ctx := context.Background()

	fmt.Println("================================================================")
	fmt.Println("Синхронизация цен и остатков 4tochki")
	fmt.Println("================================================================")
	fmt.Println()

	start := time.Now()

	// Execute sync based on type
	switch *syncType {
	case "tyres":
		logger.Info("Starting tyres prices/stock sync")
		if err := pricesStockService.SyncTyresPricesStock(ctx); err != nil {
			log.Fatalf("Failed to sync tyres prices/stock: %v", err)
		}
		printStats(ctx, pricesStockRepo, logger, "tyres")

	case "rims":
		logger.Info("Starting rims prices/stock sync")
		if err := pricesStockService.SyncRimsPricesStock(ctx); err != nil {
			log.Fatalf("Failed to sync rims prices/stock: %v", err)
		}
		printStats(ctx, pricesStockRepo, logger, "rims")

	case "all":
		logger.Info("Starting full prices/stock sync")
		if err := pricesStockService.SyncAll(ctx); err != nil {
			log.Fatalf("Failed to sync prices/stock: %v", err)
		}
		printStats(ctx, pricesStockRepo, logger, "all")

	default:
		log.Fatalf("Invalid sync type: %s (use 'tyres', 'rims', or 'all')", *syncType)
	}

	duration := time.Since(start)

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Printf("Синхронизация завершена успешно за %v\n", duration)
	fmt.Println("================================================================")
}

func printStats(ctx context.Context, repo *repository.PricesStockRepository, logger *slog.Logger, syncType string) {
	fmt.Println()
	fmt.Println("Статистика:")
	fmt.Println("-----------------------------------------------------------------")

	if syncType == "tyres" || syncType == "all" {
		uniqueCount, err := repo.CountUniqueCAEsInTyresPricesStock(ctx)
		if err != nil {
			logger.Error("Failed to count unique tyres", "error", err)
		}

		totalCount, err := repo.CountTyresPricesStock(ctx)
		if err != nil {
			logger.Error("Failed to count tyres prices/stock", "error", err)
		} else {
			fmt.Printf("Шины:\n")
			fmt.Printf("  - Уникальных товаров: %d\n", uniqueCount)
			fmt.Printf("  - Складских позиций: %d\n", totalCount)
		}
	}

	if syncType == "rims" || syncType == "all" {
		uniqueCount, err := repo.CountUniqueCAEsInRimsPricesStock(ctx)
		if err != nil {
			logger.Error("Failed to count unique rims", "error", err)
		}

		totalCount, err := repo.CountRimsPricesStock(ctx)
		if err != nil {
			logger.Error("Failed to count rims prices/stock", "error", err)
		} else {
			fmt.Printf("Диски:\n")
			fmt.Printf("  - Уникальных товаров: %d\n", uniqueCount)
			fmt.Printf("  - Складских позиций: %d\n", totalCount)
		}
	}
}
