package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/prokoleso/etalon-price-api/internal/domain"
	fourtochki "github.com/prokoleso/etalon-price-api/internal/providers/4tochki"
	"github.com/prokoleso/etalon-price-api/internal/repository"
)

const (
	tyresXMLURL = "https://b2b.4tochki.ru/export_data/Tires.xml"
	rimsXMLURL  = "https://b2b.4tochki.ru/export_data/Rims.xml"
)

// NomenclatureService handles nomenclature synchronization
type NomenclatureService struct {
	repo       *repository.NomenclatureRepository
	httpClient *http.Client
	logger     *slog.Logger
}

// NewNomenclatureService creates a new nomenclature service
func NewNomenclatureService(repo *repository.NomenclatureRepository, logger *slog.Logger) *NomenclatureService {
	return &NomenclatureService{
		repo: repo,
		httpClient: &http.Client{
			Timeout: 5 * time.Minute, // XML files can be large
		},
		logger: logger,
	}
}

// SyncTyres downloads and syncs tyres nomenclature
func (s *NomenclatureService) SyncTyres(ctx context.Context) error {
	s.logger.Info("Starting tyres nomenclature sync")

	// Create sync run
	runID, err := s.repo.CreateSyncRun(ctx, "tyres")
	if err != nil {
		return fmt.Errorf("failed to create sync run: %w", err)
	}

	// Download XML
	s.logger.Info("Downloading tyres XML", "url", tyresXMLURL)
	xmlData, err := s.downloadXML(ctx, tyresXMLURL)
	if err != nil {
		s.repo.UpdateSyncRun(ctx, runID, "failed", 0, 0, 0, err.Error())
		return fmt.Errorf("failed to download tyres XML: %w", err)
	}

	// Parse XML
	s.logger.Info("Parsing tyres XML", "size_bytes", len(xmlData))
	var root fourtochki.TyresXMLRoot
	if err := xml.Unmarshal(xmlData, &root); err != nil {
		s.repo.UpdateSyncRun(ctx, runID, "failed", 0, 0, 0, err.Error())
		return fmt.Errorf("failed to parse tyres XML: %w", err)
	}

	s.logger.Info("Parsed tyres", "count", len(root.Tyres))

	// Convert to domain models (filter only Легковая and Мотошина)
	tyres := make([]domain.NomenclatureTyre, 0, len(root.Tyres))
	filteredCount := 0
	for _, t := range root.Tyres {
		// Filter: only load "Легковая" and "Мотошина" tire types
		if t.TireType != "Легковая" && t.TireType != "Мотошина" {
			filteredCount++
			continue
		}

		tyres = append(tyres, domain.NomenclatureTyre{
			CAE:         t.CAE,
			Name:        t.Name,
			Width:       t.Width,
			Height:      t.Height,
			Diameter:    normalizeDiameter(t.Diameter),    // Normalize diameter format
			DiameterOut: normalizeDiameter(t.DiameterOut), // Normalize diameter_out format
			LoadIndex:   t.LoadIndex,
			SpeedIndex:  t.SpeedIndex,
			Model:       t.Model,
			Brand:       t.Brand,
			Season:      t.Season,
			IsStudded:   t.IsStudded,
			TireType:    t.TireType,
			RunFlat:     t.RunFlat,
			Reinforced:  t.Reinforced,
		})
	}

	s.logger.Info("Filtered tyres by type",
		"total_parsed", len(root.Tyres),
		"filtered_out", filteredCount,
		"remaining", len(tyres),
	)

	// Save to database in batches
	batchSize := 1000
	totalNew := 0
	totalSkipped := 0

	for i := 0; i < len(tyres); i += batchSize {
		end := i + batchSize
		if end > len(tyres) {
			end = len(tyres)
		}

		batch := tyres[i:end]
		newCount, skippedCount, err := s.repo.UpsertTyres(ctx, batch)
		if err != nil {
			s.repo.UpdateSyncRun(ctx, runID, "failed", len(tyres), totalNew, totalSkipped, err.Error())
			return fmt.Errorf("failed to save tyres batch: %w", err)
		}

		totalNew += newCount
		totalSkipped += skippedCount

		s.logger.Info("Saved tyres batch",
			"batch", fmt.Sprintf("%d-%d", i, end),
			"new", newCount,
			"skipped_duplicates", skippedCount,
		)
	}

	// Update sync run (using skipped as "updated" field for backwards compatibility)
	if err := s.repo.UpdateSyncRun(ctx, runID, "completed", len(tyres), totalNew, totalSkipped, ""); err != nil {
		return fmt.Errorf("failed to update sync run: %w", err)
	}

	s.logger.Info("Tyres nomenclature sync completed",
		"total", len(tyres),
		"new", totalNew,
		"skipped_duplicates", totalSkipped,
	)

	return nil
}

// SyncRims downloads and syncs rims nomenclature
func (s *NomenclatureService) SyncRims(ctx context.Context) error {
	s.logger.Info("Starting rims nomenclature sync")

	// Create sync run
	runID, err := s.repo.CreateSyncRun(ctx, "rims")
	if err != nil {
		return fmt.Errorf("failed to create sync run: %w", err)
	}

	// Download XML
	s.logger.Info("Downloading rims XML", "url", rimsXMLURL)
	xmlData, err := s.downloadXML(ctx, rimsXMLURL)
	if err != nil {
		s.repo.UpdateSyncRun(ctx, runID, "failed", 0, 0, 0, err.Error())
		return fmt.Errorf("failed to download rims XML: %w", err)
	}

	// Parse XML
	s.logger.Info("Parsing rims XML", "size_bytes", len(xmlData))
	var root fourtochki.RimsXMLRoot
	if err := xml.Unmarshal(xmlData, &root); err != nil {
		s.repo.UpdateSyncRun(ctx, runID, "failed", 0, 0, 0, err.Error())
		return fmt.Errorf("failed to parse rims XML: %w", err)
	}

	s.logger.Info("Parsed rims", "count", len(root.Rims))

	// Convert to domain models
	rims := make([]domain.NomenclatureRim, 0, len(root.Rims))
	for _, r := range root.Rims {
		rims = append(rims, domain.NomenclatureRim{
			CAE:           r.CAE,
			Name:          r.Name,
			Width:         r.Width,
			Diameter:      r.Diameter,
			BoltsCount:    r.BoltsCount,
			BoltsSpacing:  r.BoltsSpacing,
			BoltsSpacing2: r.BoltsSpacing2,
			ET:            r.ET,
			DIA:           r.DIA,
			Model:         r.Model,
			Brand:         r.Brand,
			Color:         r.Color,
			RimType:       r.RimType,
		})
	}

	// Save to database in batches
	batchSize := 1000
	totalNew := 0
	totalSkipped := 0

	for i := 0; i < len(rims); i += batchSize {
		end := i + batchSize
		if end > len(rims) {
			end = len(rims)
		}

		batch := rims[i:end]
		newCount, skippedCount, err := s.repo.UpsertRims(ctx, batch)
		if err != nil {
			s.repo.UpdateSyncRun(ctx, runID, "failed", len(rims), totalNew, totalSkipped, err.Error())
			return fmt.Errorf("failed to save rims batch: %w", err)
		}

		totalNew += newCount
		totalSkipped += skippedCount

		s.logger.Info("Saved rims batch",
			"batch", fmt.Sprintf("%d-%d", i, end),
			"new", newCount,
			"skipped_duplicates", skippedCount,
		)
	}

	// Update sync run (using skipped as "updated" field for backwards compatibility)
	if err := s.repo.UpdateSyncRun(ctx, runID, "completed", len(rims), totalNew, totalSkipped, ""); err != nil {
		return fmt.Errorf("failed to update sync run: %w", err)
	}

	s.logger.Info("Rims nomenclature sync completed",
		"total", len(rims),
		"new", totalNew,
		"skipped_duplicates", totalSkipped,
	)

	return nil
}

// SyncAll syncs both tyres and rims
func (s *NomenclatureService) SyncAll(ctx context.Context) error {
	if err := s.SyncTyres(ctx); err != nil {
		return fmt.Errorf("failed to sync tyres: %w", err)
	}

	if err := s.SyncRims(ctx); err != nil {
		return fmt.Errorf("failed to sync rims: %w", err)
	}

	return nil
}

// downloadXML downloads XML file from URL
func (s *NomenclatureService) downloadXML(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return data, nil
}

// normalizeDiameter normalizes diameter format
// R13,00 → R13
// ZR18,00 → ZR18
// R22,50C → R22.5C
// R16,50 → R16.5
func normalizeDiameter(diameter string) string {
	if diameter == "" {
		return diameter
	}

	// Replace comma with dot first
	diameter = strings.Replace(diameter, ",", ".", -1)

	// Remove trailing ".00" (R13.00 → R13)
	if len(diameter) >= 3 && diameter[len(diameter)-3:] == ".00" {
		return diameter[:len(diameter)-3]
	}

	// Remove trailing ".00C" (R22.00C → R22C)
	if len(diameter) >= 4 && diameter[len(diameter)-4:] == ".00C" {
		return diameter[:len(diameter)-4] + "C"
	}

	// Remove trailing "0" from decimals (R16.50 → R16.5, R22.50C → R22.5C)
	if len(diameter) >= 2 && diameter[len(diameter)-1] == '0' {
		// Check if previous char is a digit (not C or letter)
		prevChar := diameter[len(diameter)-2]
		if prevChar >= '0' && prevChar <= '9' {
			// Check if there's a dot before
			for i := len(diameter) - 3; i >= 0; i-- {
				if diameter[i] == '.' {
					return diameter[:len(diameter)-1]
				}
				if diameter[i] < '0' || diameter[i] > '9' {
					break
				}
			}
		}
	}

	// Remove trailing "0C" from decimals (R22.50C already handled, but for R16.50C → R16.5C)
	if len(diameter) >= 3 && diameter[len(diameter)-2:] == "0C" {
		prevChar := diameter[len(diameter)-3]
		if prevChar >= '0' && prevChar <= '9' {
			return diameter[:len(diameter)-2] + "C"
		}
	}

	return diameter
}
