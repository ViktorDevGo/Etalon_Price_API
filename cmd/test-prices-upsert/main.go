package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/config"
	"github.com/prokoleso/etalon-price-api/internal/domain"
	"github.com/prokoleso/etalon-price-api/internal/repository"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()
	repo := repository.NewPricesStockRepository(pool)

	fmt.Println("================================================================")
	fmt.Println("Тест логики UPSERT для таблиц цен и остатков")
	fmt.Println("================================================================")
	fmt.Println()

	// Тестовые данные
	testCAE := "TEST_CAE_12345"
	testWarehouse := "Тестовый склад"
	testProvider := "Форточки"

	// Шаг 1: Удаляем тестовую запись если она есть
	fmt.Println("📋 Шаг 1: Очистка тестовых данных...")
	_, err = pool.Exec(ctx, "DELETE FROM tyres_prices_stock WHERE cae = $1", testCAE)
	if err != nil {
		log.Fatalf("Failed to cleanup: %v", err)
	}
	fmt.Println("✅ Тестовые данные очищены")
	fmt.Println()

	// Шаг 2: Вставляем новую запись
	fmt.Println("📋 Шаг 2: Вставка НОВОЙ записи (isimport=1)...")
	newItems := []domain.TyrePriceStock{
		{
			CAE:           testCAE,
			WarehouseName: testWarehouse,
			Price:         10000, // 100.00 руб
			Stock:         50,
			IsImport:      1,
			Provider:      testProvider,
		},
	}

	newCount, updatedCount, err := repo.UpsertTyresPricesStock(ctx, newItems)
	if err != nil {
		log.Fatalf("Failed to insert: %v", err)
	}

	fmt.Printf("✅ Результат: new=%d, updated=%d\n", newCount, updatedCount)
	
	// Проверяем что запись добавилась
	var price, stock, isimport int
	err = pool.QueryRow(ctx, `
		SELECT price, stock, isimport 
		FROM tyres_prices_stock 
		WHERE cae = $1 AND warehouse_name = $2 AND provider = $3
	`, testCAE, testWarehouse, testProvider).Scan(&price, &stock, &isimport)
	
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	
	fmt.Printf("   📊 Данные в БД: price=%d, stock=%d, isimport=%d\n", price, stock, isimport)
	
	if isimport != 1 {
		log.Fatalf("❌ ОШИБКА: При INSERT isimport должен быть 1, получено %d", isimport)
	}
	fmt.Println("   ✅ isimport=1 (корректно для новой записи)")
	fmt.Println()

	// Шаг 3: Обновляем существующую запись
	fmt.Println("📋 Шаг 3: Обновление СУЩЕСТВУЮЩЕЙ записи...")
	updateItems := []domain.TyrePriceStock{
		{
			CAE:           testCAE,
			WarehouseName: testWarehouse,
			Price:         15000, // 150.00 руб (изменилась цена)
			Stock:         75,     // изменился остаток
			IsImport:      1,      // в исходных данных 1, но должно стать 0
			Provider:      testProvider,
		},
	}

	newCount, updatedCount, err = repo.UpsertTyresPricesStock(ctx, updateItems)
	if err != nil {
		log.Fatalf("Failed to update: %v", err)
	}

	fmt.Printf("✅ Результат: new=%d, updated=%d\n", newCount, updatedCount)
	
	// Проверяем что запись обновилась
	err = pool.QueryRow(ctx, `
		SELECT price, stock, isimport 
		FROM tyres_prices_stock 
		WHERE cae = $1 AND warehouse_name = $2 AND provider = $3
	`, testCAE, testWarehouse, testProvider).Scan(&price, &stock, &isimport)
	
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	
	fmt.Printf("   📊 Данные в БД: price=%d, stock=%d, isimport=%d\n", price, stock, isimport)
	
	if price != 15000 {
		log.Fatalf("❌ ОШИБКА: price должен быть 15000, получено %d", price)
	}
	if stock != 75 {
		log.Fatalf("❌ ОШИБКА: stock должен быть 75, получено %d", stock)
	}
	if isimport != 0 {
		log.Fatalf("❌ ОШИБКА: При UPDATE isimport должен стать 0, получено %d", isimport)
	}
	
	fmt.Println("   ✅ Цена обновлена: 10000 → 15000")
	fmt.Println("   ✅ Остаток обновлен: 50 → 75")
	fmt.Println("   ✅ isimport=0 (корректно для обновленной записи)")
	fmt.Println()

	// Шаг 4: Еще раз обновляем (проверяем что isimport остается 0)
	fmt.Println("📋 Шаг 4: Повторное обновление...")
	updateItems2 := []domain.TyrePriceStock{
		{
			CAE:           testCAE,
			WarehouseName: testWarehouse,
			Price:         20000,
			Stock:         100,
			IsImport:      1, // в данных 1, но должно остаться 0
			Provider:      testProvider,
		},
	}

	newCount, updatedCount, err = repo.UpsertTyresPricesStock(ctx, updateItems2)
	if err != nil {
		log.Fatalf("Failed to update: %v", err)
	}

	err = pool.QueryRow(ctx, `
		SELECT price, stock, isimport 
		FROM tyres_prices_stock 
		WHERE cae = $1 AND warehouse_name = $2 AND provider = $3
	`, testCAE, testWarehouse, testProvider).Scan(&price, &stock, &isimport)
	
	if err != nil {
		log.Fatalf("Failed to query: %v", err)
	}
	
	fmt.Printf("   📊 Данные в БД: price=%d, stock=%d, isimport=%d\n", price, stock, isimport)
	
	if isimport != 0 {
		log.Fatalf("❌ ОШИБКА: isimport должен оставаться 0, получено %d", isimport)
	}
	
	fmt.Println("   ✅ isimport=0 (остался 0 при повторном обновлении)")
	fmt.Println()

	// Очистка
	fmt.Println("📋 Очистка тестовых данных...")
	_, err = pool.Exec(ctx, "DELETE FROM tyres_prices_stock WHERE cae = $1", testCAE)
	if err != nil {
		log.Fatalf("Failed to cleanup: %v", err)
	}
	
	fmt.Println("================================================================")
	fmt.Println("✅ ВСЕ ТЕСТЫ ПРОЙДЕНЫ УСПЕШНО!")
	fmt.Println("================================================================")
	fmt.Println()
	fmt.Println("Логика работает корректно:")
	fmt.Println("  • При INSERT новой записи → isimport берется из данных")
	fmt.Println("  • При UPDATE существующей записи → isimport = 0")
	fmt.Println("  • При повторном UPDATE → isimport остается 0")
}
