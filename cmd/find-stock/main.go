package main

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	pool, _ := pgxpool.New(context.Background(), os.Getenv("DATABASE_DSN"))
	defer pool.Close()

	fmt.Println("🔍 Поиск CAE в tyres_prices_stock с остатком > 10...\n")

	rows, _ := pool.Query(context.Background(), `
		SELECT DISTINCT tps.cae, nt.name, SUM(tps.stock) as total_stock
		FROM tyres_prices_stock tps
		LEFT JOIN nomenclature_tyres nt ON tps.cae = nt.cae
		WHERE tps.stock > 10
		GROUP BY tps.cae, nt.name
		ORDER BY total_stock DESC
		LIMIT 10
	`)
	defer rows.Close()

	fmt.Println("CAE с самыми большими остатками:")
	for rows.Next() {
		var cae, name string
		var stock int
		rows.Scan(&cae, &name, &stock)
		fmt.Printf("  ✅ %s - %s (остаток: %d шт)\n", cae, name, stock)
	}
}
