package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/prokoleso/etalon-price-api/internal/config"
	fourtochki "github.com/prokoleso/etalon-price-api/internal/providers/4tochki"
)

func main() {
	// Parse flags
	codes := flag.String("codes", "", "Comma-separated list of product codes to test")
	method := flag.String("method", "GetGoodsInfo", "SOAP method to test (GetGoodsInfo or GetGoodsPriceRestByCode)")
	flag.Parse()

	if *codes == "" {
		log.Fatal("Error: --codes parameter is required\n\nUsage: ./bin/test-soap --codes=CODE1,CODE2 [--method=GetGoodsInfo]")
	}

	// Parse codes
	codeList := strings.Split(*codes, ",")
	for i := range codeList {
		codeList[i] = strings.TrimSpace(codeList[i])
	}

	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("4tochki SOAP API Test Utility")
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println()

	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Printf("Configuration:\n")
	fmt.Printf("  WSDL URL: %s\n", cfg.Fourtochki.WSDLURL)
	fmt.Printf("  Login:    %s\n", cfg.Fourtochki.Login)
	fmt.Printf("  Password: %s\n", maskPassword(cfg.Fourtochki.Password))
	fmt.Printf("  Method:   %s\n", *method)
	fmt.Printf("  Codes:    %v\n", codeList)
	fmt.Println()

	// Create client with detailed logging
	client := fourtochki.NewClient(fourtochki.ClientConfig{
		WSDLURL:    cfg.Fourtochki.WSDLURL,
		Login:      cfg.Fourtochki.Login,
		Password:   cfg.Fourtochki.Password,
		Timeout:    30 * time.Second,
		RetryCount: 0, // No retries for testing
		RetryDelay: 0,
		Logger:     nil, // Will use default logger
	})

	ctx := context.Background()

	// Execute based on method
	switch *method {
	case "GetGoodsInfo":
		testGetGoodsInfo(ctx, client, codeList)
	case "GetGoodsPriceRestByCode":
		testGetGoodsPriceRestByCode(ctx, client, codeList)
	default:
		log.Fatalf("Unknown method: %s. Use GetGoodsInfo or GetGoodsPriceRestByCode", *method)
	}
}

func testGetGoodsInfo(ctx context.Context, client *fourtochki.Client, codes []string) {
	fmt.Println("Testing GetGoodsInfo...")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println()

	start := time.Now()
	resp, err := client.GetGoodsInfo(ctx, codes)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ ERROR: %v\n", err)
		fmt.Printf("Duration: %v\n", duration)
		return
	}

	fmt.Printf("✅ Request successful (took %v)\n", duration)
	fmt.Println()

	// Check for API error
	if resp.Result.Error != nil {
		fmt.Println("⚠️  API Error:")
		fmt.Printf("  Code:    %d\n", resp.Result.Error.Code)
		fmt.Printf("  Comment: %s\n", resp.Result.Error.Comment)
		fmt.Println()
	}

	// Show results
	fmt.Printf("Results:\n")
	fmt.Printf("  Tyres: %d items\n", len(resp.Result.TyreList))
	fmt.Printf("  Rims:  %d items\n", len(resp.Result.RimList))
	fmt.Println()

	// Show tyres
	if len(resp.Result.TyreList) > 0 {
		fmt.Println("Tyres:")
		for i, tyre := range resp.Result.TyreList {
			fmt.Printf("\n  [%d] %s\n", i+1, tyre.Code)
			fmt.Printf("      Brand:       %s\n", tyre.Brand)
			fmt.Printf("      Model:       %s\n", tyre.Model)
			fmt.Printf("      Size:        %d/%d R%d\n", tyre.Width, tyre.Height, tyre.Diameter)
			fmt.Printf("      Season:      %s\n", tyre.Season)
			fmt.Printf("      Load/Speed:  %s/%s\n", tyre.LoadIndex, tyre.SpeedIndex)
			fmt.Printf("      Studded:     %s\n", tyre.Thorns)
			fmt.Printf("      RunFlat:     %s\n", tyre.RunFlat)
			fmt.Printf("      Description: %s\n", tyre.Description)
		}
	}

	// Show rims
	if len(resp.Result.RimList) > 0 {
		fmt.Println("Rims:")
		for i, rim := range resp.Result.RimList {
			fmt.Printf("\n  [%d] %s\n", i+1, rim.Code)
			fmt.Printf("      Brand:        %s\n", rim.Brand)
			fmt.Printf("      Model:        %s\n", rim.Model)
			fmt.Printf("      Size:         %.1fx%.1f\n", rim.DiskWidth, rim.DiskDiameter)
			fmt.Printf("      Bolt Pattern: %s\n", rim.BoltPattern)
			fmt.Printf("      Offset:       %d\n", rim.Offset)
			fmt.Printf("      Center Bore:  %.1f\n", rim.CenterBore)
			fmt.Printf("      Color:        %s\n", rim.Color)
			fmt.Printf("      Description:  %s\n", rim.Description)
		}
	}

	// Show raw response as JSON for debugging
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("Raw Response (JSON):")
	fmt.Println(strings.Repeat("-", 80))
	jsonData, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(jsonData))
}

func testGetGoodsPriceRestByCode(ctx context.Context, client *fourtochki.Client, codes []string) {
	fmt.Println("Testing GetGoodsPriceRestByCode...")
	fmt.Println(strings.Repeat("-", 80))
	fmt.Println()

	start := time.Now()
	resp, err := client.GetGoodsPriceRestByCode(ctx, codes)
	duration := time.Since(start)

	if err != nil {
		fmt.Printf("❌ ERROR: %v\n", err)
		fmt.Printf("Duration: %v\n", duration)
		return
	}

	fmt.Printf("✅ Request successful (took %v)\n", duration)
	fmt.Println()

	// Check for API error
	if resp.Result.Error != nil {
		fmt.Println("⚠️  API Error:")
		fmt.Printf("  Code:    %d\n", resp.Result.Error.Code)
		fmt.Printf("  Comment: %s\n", resp.Result.Error.Comment)
		fmt.Println()
	}

	// Show results
	fmt.Printf("Results:\n")
	fmt.Printf("  Price/Stock Items: %d\n", len(resp.Result.Items))
	fmt.Println()

	// Show price/stock items
	if len(resp.Result.Items) > 0 {
		fmt.Println("Price & Stock Information:")
		for i, item := range resp.Result.Items {
			fmt.Printf("\n  [%d] %s\n", i+1, item.Code)
			fmt.Printf("      Price:         %.2f %s\n", item.Price, item.Currency)
			fmt.Printf("      Quantity:      %d\n", item.Quantity)
			fmt.Printf("      Warehouse:     %s (%s)\n", item.WarehouseName, item.WarehouseCode)
			fmt.Printf("      Delivery Days: %d\n", item.DeliveryDays)
			fmt.Printf("      Min Order:     %d\n", item.MinOrder)
			fmt.Printf("      Available:     %s\n", item.Available)
			fmt.Printf("      Stock Status:  %s\n", item.Stock)
			fmt.Printf("      Region:        %s\n", item.Region)
			fmt.Printf("      Updated:       %s\n", item.Updated)
		}
	}

	// Show raw response as JSON for debugging
	fmt.Println()
	fmt.Println(strings.Repeat("=", 80))
	fmt.Println("Raw Response (JSON):")
	fmt.Println(strings.Repeat("-", 80))
	jsonData, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(jsonData))
}

func maskPassword(password string) string {
	if len(password) <= 2 {
		return "***"
	}
	return password[:2] + strings.Repeat("*", len(password)-2)
}
