package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
	fourtochki "github.com/prokoleso/etalon-price-api/internal/providers/4tochki"
	"github.com/prokoleso/etalon-price-api/internal/repository"
	"github.com/prokoleso/etalon-price-api/internal/service"
)

func main() {
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

	// Create repository and service
	warehouseRepo := repository.NewWarehouseRepository(pool)
	warehouseService := service.NewWarehouseService(warehouseRepo, fourtochkiClient, logger)

	ctx := context.Background()

	fmt.Println("================================================================")
	fmt.Println("Синхронизация складов 4tochki")
	fmt.Println("================================================================")
	fmt.Println()

	logger.Info("Starting warehouses sync")
	if err := warehouseService.SyncWarehouses(ctx); err != nil {
		log.Fatalf("Failed to sync warehouses: %v", err)
	}

	// Print stats
	count, err := warehouseRepo.CountWarehouses(ctx)
	if err != nil {
		logger.Error("Failed to count warehouses", "error", err)
	} else {
		fmt.Println()
		fmt.Println("Статистика:")
		fmt.Println("-----------------------------------------------------------------")
		fmt.Printf("Всего складов в БД: %d\n", count)
	}

	// List warehouses
	warehouses, err := warehouseRepo.GetAllWarehouses(ctx)
	if err != nil {
		logger.Error("Failed to get warehouses", "error", err)
	} else if len(warehouses) > 0 {
		fmt.Println()
		fmt.Println("Список складов:")
		fmt.Println("-----------------------------------------------------------------")
		for _, w := range warehouses {
			deliveryStr := ""
			if w.HaveDelivery {
				deliveryStr = "доставка"
			}
			if w.HavePickup {
				if deliveryStr != "" {
					deliveryStr += ", "
				}
				deliveryStr += "самовывоз"
			}
			fmt.Printf("  [%d] %s (%s) - %s\n", w.ID, w.Name, w.ShortName, deliveryStr)
		}
	}

	fmt.Println()
	fmt.Println("================================================================")
	fmt.Println("Синхронизация завершена успешно")
	fmt.Println("================================================================")
}
