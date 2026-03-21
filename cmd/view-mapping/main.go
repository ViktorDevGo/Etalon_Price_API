package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbDSN := os.Getenv("DATABASE_DSN")
	if dbDSN == "" {
		log.Fatal("DATABASE_DSN must be set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbDSN)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	fmt.Println("\n📊 Severavto Tyres Mapping Statistics\n")

	// Total count
	var total int
	pool.QueryRow(ctx, "SELECT COUNT(*) FROM severavto_tyres_mapping").Scan(&total)
	fmt.Printf("Total mappings: %d\n\n", total)

	// By priority
	fmt.Println("By priority:")
	rows, err := pool.Query(ctx, `
		SELECT match_priority, COUNT(*) as cnt
		FROM severavto_tyres_mapping
		GROUP BY match_priority
		ORDER BY match_priority
	`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var priority, count int
		rows.Scan(&priority, &count)
		var desc string
		switch priority {
		case 1:
			desc = "Exact (brand+model+sizes+indices)"
		case 2:
			desc = "Brand+model+width+indices"
		case 3:
			desc = "Brand+width+indices"
		case 4:
			desc = "Brand+sizes"
		}
		fmt.Printf("   %d - %s: %d (%.1f%%)\n", priority, desc, count, float64(count)/float64(total)*100)
	}

	// Sample mappings by priority
	fmt.Println("\n📝 Sample mappings (5 per priority):\n")
	for priority := 1; priority <= 4; priority++ {
		fmt.Printf("Priority %d:\n", priority)
		rows, err := pool.Query(ctx, `
			SELECT m.cae, m.commodity_id, m.match_details,
			       nt.brand, nt.model, nt.width, nt.height, nt.diameter,
			       nts.brand, nts.model, nts.width, nts.height, nts.diameter
			FROM severavto_tyres_mapping m
			INNER JOIN nomenclature_tyres nt ON m.cae = nt.cae
			INNER JOIN nomenclature_tyres_severavto nts ON m.commodity_id = nts.commodity_id
			WHERE m.match_priority = $1
			LIMIT 5
		`, priority)
		if err != nil {
			log.Fatal(err)
		}

		for rows.Next() {
			var cae, commodityID, detailsJSON string
			var ftBrand, ftModel, svBrand, svModel string
			var ftWidth, ftHeight, svWidth, svHeight float64
			var ftDiam, svDiam string

			rows.Scan(&cae, &commodityID, &detailsJSON,
				&ftBrand, &ftModel, &ftWidth, &ftHeight, &ftDiam,
				&svBrand, &svModel, &svWidth, &svHeight, &svDiam)

			fmt.Printf("   CAE: %s → commodity_id: %s\n", cae, commodityID)
			fmt.Printf("      Форточки:  %s %s %.0f/%.0f R%s\n", ftBrand, ftModel, ftWidth, ftHeight, ftDiam)
			fmt.Printf("      Северавто: %s %s %.0f/%.0f R%s\n", svBrand, svModel, svWidth, svHeight, svDiam)
			fmt.Println()
		}
		rows.Close()
	}

	fmt.Println("\n✅ Mapping ready for use in price/stock sync!")
	fmt.Println("   Use this mapping to import Severavto prices into tyres_prices_stock")
}
