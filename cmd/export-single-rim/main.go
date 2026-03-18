package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dimensionsByDiameter = map[float64]struct {
	Weight, Width, Height, Length float64
}{
	14: {6.5, 40, 40, 18},
	15: {8.0, 42, 42, 18},
	16: {9.5, 44, 44, 20},
	17: {11.5, 46, 46, 22},
	18: {13.0, 48, 48, 23},
	19: {15.5, 50, 50, 24},
	20: {17.0, 52, 52, 25},
	21: {19.5, 55, 55, 28},
	22: {23.0, 58, 58, 30},
}

func main() {
	caeToFind := "1151011"
	if len(os.Args) > 1 {
		caeToFind = os.Args[1]
	}

	dsn := os.Getenv("DATABASE_DSN")
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer pool.Close()

	var cae, name string
	var diameter float64
	err = pool.QueryRow(context.Background(),
		"SELECT cae, name, diameter FROM nomenclature_rims WHERE cae = $1",
		caeToFind).Scan(&cae, &name, &diameter)
	if err != nil {
		log.Fatalf("CAE %s не найден: %v", caeToFind, err)
	}

	outputPath := fmt.Sprintf("/Users/viktor/Desktop/rims_test_%s.csv", caeToFind)
	file, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	w := csv.NewWriter(file)
	defer w.Flush()

	w.Write([]string{"IE_XML_ID", "IE_NAME", "CP_WEIGHT", "CP_WIDTH", "CP_HEIGHT", "CP_LENGTH"})

	dims, ok := dimensionsByDiameter[diameter]
	if !ok {
		dims = struct{ Weight, Width, Height, Length float64 }{8.0, 42, 42, 18}
	}

	w.Write([]string{
		cae,
		name,
		fmt.Sprintf("%.0f", dims.Weight*1000),
		fmt.Sprintf("%.0f", dims.Width*10),
		fmt.Sprintf("%.0f", dims.Height*10),
		fmt.Sprintf("%.0f", dims.Length*10),
	})

	fmt.Printf("✅ CAE: %s\n", cae)
	fmt.Printf("   Name: %s\n", name)
	fmt.Printf("   Diameter: %.0f\"\n", diameter)
	fmt.Printf("   Weight: %.0f г\n", dims.Weight*1000)
	fmt.Printf("   Dimensions: %.0f×%.0f×%.0f мм\n", dims.Width*10, dims.Height*10, dims.Length*10)
	fmt.Printf("📄 File: %s\n", outputPath)
}
