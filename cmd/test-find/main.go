package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
	fourtochki "github.com/prokoleso/etalon-price-api/internal/providers/4tochki"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("=================================================================")
	fmt.Println("Тестирование методов GetFindTyre и GetFindDisk")
	fmt.Println("=================================================================")
	fmt.Println()

	client := fourtochki.NewClient(fourtochki.ClientConfig{
		WSDLURL:    cfg.Fourtochki.WSDLURL,
		Login:      cfg.Fourtochki.Login,
		Password:   cfg.Fourtochki.Password,
		Timeout:    30 * time.Second,
		RetryCount: 0,
		RetryDelay: 0,
		Logger:     nil,
	})

	ctx := context.Background()

	// Test GetFindTyre
	fmt.Println("Запрос GetFindTyre (page=0, pageSize=100)...")
	fmt.Println("-----------------------------------------------------------------")
	start := time.Now()
	tyreResp, err := client.GetFindTyre(ctx, 0, 100)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ ОШИБКА: %v\n", err)
		fmt.Printf("Длительность: %v\n", duration)
	} else {
		fmt.Printf("✅ Успешно (заняло %v)\n", duration)

		if tyreResp.Result.Error != nil && (tyreResp.Result.Error.Code != 0 || tyreResp.Result.Error.Comment != "") {
			fmt.Printf("⚠️  Ошибка API:\n")
			fmt.Printf("  Код:    %d\n", tyreResp.Result.Error.Code)
			fmt.Printf("  Текст:  %s\n", tyreResp.Result.Error.Comment)
		} else {
			fmt.Printf("Получено кодов шин: %d\n", len(tyreResp.Result.Items))

			if len(tyreResp.Result.Items) > 0 {
				fmt.Println("\nПримеры товаров (первые 3):")
				limit := 3
				if len(tyreResp.Result.Items) < limit {
					limit = len(tyreResp.Result.Items)
				}
				for i := 0; i < limit; i++ {
					item := tyreResp.Result.Items[i]
					fmt.Printf("\n  %d. Код: %s\n", i+1, item.Code)
					fmt.Printf("     Марка: %s\n", item.Marka)
					fmt.Printf("     Модель: %s\n", item.Model)
					fmt.Printf("     Название: %s\n", item.Name)
					fmt.Printf("     Сезон: %s, Шипы: %s, Тип: %s\n", item.Season, item.Thorn, item.Type)
					if len(item.Whpr) > 0 {
						fmt.Printf("     Склады (%d):\n", len(item.Whpr))
						for j, wh := range item.Whpr {
							if j < 2 { // Show first 2 warehouses
								fmt.Printf("       - Склад %d: цена %.2f руб, остаток %d шт\n",
									wh.Wrh, float64(wh.Price)/100, wh.Rest)
							}
						}
					}
				}
			}
		}
	}

	fmt.Println()
	fmt.Println()

	// Test GetFindDisk
	fmt.Println("Запрос GetFindDisk (page=0, pageSize=100)...")
	fmt.Println("-----------------------------------------------------------------")
	start = time.Now()
	diskResp, err := client.GetFindDisk(ctx, 0, 100)
	duration = time.Since(start)

	if err != nil {
		fmt.Printf("❌ ОШИБКА: %v\n", err)
		fmt.Printf("Длительность: %v\n", duration)
	} else {
		fmt.Printf("✅ Успешно (заняло %v)\n", duration)

		if diskResp.Result.Error != nil && (diskResp.Result.Error.Code != 0 || diskResp.Result.Error.Comment != "") {
			fmt.Printf("⚠️  Ошибка API:\n")
			fmt.Printf("  Код:    %d\n", diskResp.Result.Error.Code)
			fmt.Printf("  Текст:  %s\n", diskResp.Result.Error.Comment)
		} else {
			fmt.Printf("Получено кодов дисков: %d\n", len(diskResp.Result.Items))

			if len(diskResp.Result.Items) > 0 {
				fmt.Println("\nПримеры товаров (первые 3):")
				limit := 3
				if len(diskResp.Result.Items) < limit {
					limit = len(diskResp.Result.Items)
				}
				for i := 0; i < limit; i++ {
					item := diskResp.Result.Items[i]
					fmt.Printf("\n  %d. Код: %s\n", i+1, item.Code)
					fmt.Printf("     Марка: %s\n", item.Marka)
					fmt.Printf("     Модель: %s\n", item.Model)
					fmt.Printf("     Название: %s\n", item.Name)
					fmt.Printf("     Тип: %s\n", item.Type)
					if len(item.Whpr) > 0 {
						fmt.Printf("     Склады (%d):\n", len(item.Whpr))
						for j, wh := range item.Whpr {
							if j < 2 { // Show first 2 warehouses
								fmt.Printf("       - Склад %d: цена %.2f руб, остаток %d шт\n",
									wh.Wrh, float64(wh.Price)/100, wh.Rest)
							}
						}
					}
				}
			}
		}
	}

	fmt.Println()
	fmt.Println("=================================================================")
	fmt.Println("Тестирование завершено")
	fmt.Println("=================================================================")
}
