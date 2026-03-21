package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/prokoleso/etalon-price-api/internal/providers/severavto"
)

func main() {
	outputDir := flag.String("output-dir", ".", "Output directory for CSV files")
	flag.Parse()

	// Load environment variables
	_ = godotenv.Load()

	baseURL := os.Getenv("SEVERAVTO_BASE_URL")
	apiKey := os.Getenv("SEVERAVTO_API_KEY")

	if baseURL == "" || apiKey == "" {
		log.Fatal("SEVERAVTO_BASE_URL and SEVERAVTO_API_KEY must be set")
	}

	// Create Severavto client
	client := severavto.NewClient(severavto.ClientConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Timeout: 120 * time.Second,
	})

	ctx := context.Background()

	// Fetch tyres
	fmt.Println("📥 Fetching tyres from Severavto API...")
	tyresData, err := client.FetchTyres(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch tyres: %v", err)
	}

	// Fetch rims
	fmt.Println("📥 Fetching rims from Severavto API...")
	rimsData, err := client.FetchRims(ctx)
	if err != nil {
		log.Fatalf("Failed to fetch rims: %v", err)
	}

	// Export tyres
	tyresFile := fmt.Sprintf("%s/Северавто_Шины_Фото_%s.csv",
		*outputDir, time.Now().Format("20060102"))
	if err := exportTyres(tyresData, tyresFile); err != nil {
		log.Fatalf("Failed to export tyres: %v", err)
	}
	fmt.Printf("✅ Tyres exported: %s\n", tyresFile)

	// Export rims
	rimsFile := fmt.Sprintf("%s/Северавто_Диски_Фото_%s.csv",
		*outputDir, time.Now().Format("20060102"))
	if err := exportRims(rimsData, rimsFile); err != nil {
		log.Fatalf("Failed to export rims: %v", err)
	}
	fmt.Printf("✅ Rims exported: %s\n", rimsFile)

	fmt.Println("\n📊 Summary:")
	fmt.Printf("Tyres with pictures: %d\n", countPictures(tyresData.Commodities))
	fmt.Printf("Rims with pictures: %d\n", countPictures(rimsData.Commodities))
}

func exportTyres(data *severavto.CommoditiesXML, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write UTF-8 BOM for Excel
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()

	// Header
	header := []string{
		"ID товара",
		"Название",
		"Бренд",
		"Модель",
		"Размер",
		"Сезон",
		"Шипы",
		"Индекс нагрузки",
		"Индекс скорости",
		"URL фото",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Group by commodity_id to get unique items
	uniqueItems := make(map[string]*severavto.Commodity)
	for _, item := range data.Commodities {
		if _, exists := uniqueItems[item.ID]; !exists {
			uniqueItems[item.ID] = &item
		}
	}

	// Write data
	for _, item := range uniqueItems {
		size := fmt.Sprintf("%s/%s R%s", item.Width, item.Height, item.Diameter)

		row := []string{
			item.ID,
			item.ModifName,
			item.Brand,
			item.Model,
			size,
			item.Season,
			item.Thorning,
			item.Load,
			item.Speed,
			item.Picture,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func exportRims(data *severavto.CommoditiesXML, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// Write UTF-8 BOM for Excel
	file.Write([]byte{0xEF, 0xBB, 0xBF})

	writer := csv.NewWriter(file)
	writer.Comma = ';'
	defer writer.Flush()

	// Header
	header := []string{
		"ID товара",
		"Название",
		"Бренд",
		"Модель",
		"Размер",
		"URL фото",
	}
	if err := writer.Write(header); err != nil {
		return err
	}

	// Group by commodity_id to get unique items
	uniqueItems := make(map[string]*severavto.Commodity)
	for _, item := range data.Commodities {
		if _, exists := uniqueItems[item.ID]; !exists {
			uniqueItems[item.ID] = &item
		}
	}

	// Write data
	for _, item := range uniqueItems {
		size := fmt.Sprintf("%sx%s", item.Width, item.Diameter)

		row := []string{
			item.ID,
			item.ModifName,
			item.Brand,
			item.Model,
			size,
			item.Picture,
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}

	return nil
}

func countPictures(commodities []severavto.Commodity) int {
	uniqueWithPictures := make(map[string]bool)
	for _, item := range commodities {
		if item.Picture != "" {
			uniqueWithPictures[item.ID] = true
		}
	}
	return len(uniqueWithPictures)
}
