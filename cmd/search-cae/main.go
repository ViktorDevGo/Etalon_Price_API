package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	search := "1151011"
	if len(os.Args) > 1 {
		search = os.Args[1]
	}

	pool, _ := pgxpool.New(context.Background(), os.Getenv("DATABASE_DSN"))
	defer pool.Close()

	fmt.Printf("🔍 Поиск CAE '%s'...\n\n", search)

	// Точное совпадение
	rows, _ := pool.Query(context.Background(),
		"SELECT cae, name FROM nomenclature_rims WHERE cae = $1 LIMIT 5", search)
	count := 0
	for rows.Next() {
		var cae, name string
		rows.Scan(&cae, &name)
		fmt.Printf("✅ Найдено: %s - %s\n", cae, name)
		count++
	}
	rows.Close()

	if count == 0 {
		// Частичное совпадение
		rows, _ = pool.Query(context.Background(),
			"SELECT cae, name FROM nomenclature_rims WHERE cae LIKE $1 LIMIT 10", "%"+search+"%")
		for rows.Next() {
			var cae, name string
			rows.Scan(&cae, &name)
			fmt.Printf("  %s - %s\n", cae, name)
			count++
		}
		rows.Close()
	}

	if count == 0 {
		fmt.Println("❌ Не найдено")
		fmt.Println("\n📋 Первые 5 CAE из базы:")
		rows, _ = pool.Query(context.Background(),
			"SELECT cae, name FROM nomenclature_rims ORDER BY id LIMIT 5")
		for rows.Next() {
			var cae, name string
			rows.Scan(&cae, &name)
			fmt.Printf("  %s - %s\n", cae, name)
		}
		rows.Close()
	}
}
