package service

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/prokoleso/etalon-price-api/internal/domain"
	fourtochki "github.com/prokoleso/etalon-price-api/internal/providers/4tochki"
)

// SyncTyresWithResult downloads and syncs tyres nomenclature and returns detailed statistics
func (s *NomenclatureService) SyncTyresWithResult(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		Type:            "tyres",
		StartedAt:       time.Now(),
		FilteredOutType: "Грузовая, Спецшина",
	}

	s.logger.Info("Starting tyres nomenclature sync")

	// Create sync run
	runID, err := s.repo.CreateSyncRun(ctx, "tyres")
	if err != nil {
		result.Error = err
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
		return result, fmt.Errorf("failed to create sync run: %w", err)
	}

	// Download XML
	s.logger.Info("Downloading tyres XML", "url", tyresXMLURL)
	xmlData, err := s.downloadXML(ctx, tyresXMLURL)
	if err != nil {
		s.repo.UpdateSyncRun(ctx, runID, "failed", 0, 0, 0, err.Error())
		result.Error = err
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
		return result, fmt.Errorf("failed to download tyres XML: %w", err)
	}

	// Parse XML
	s.logger.Info("Parsing tyres XML", "size_bytes", len(xmlData))
	var root fourtochki.TyresXMLRoot
	if err := xml.Unmarshal(xmlData, &root); err != nil {
		s.repo.UpdateSyncRun(ctx, runID, "failed", 0, 0, 0, err.Error())
		result.Error = err
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
		return result, fmt.Errorf("failed to parse tyres XML: %w", err)
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
			Diameter:    normalizeDiameter(t.Diameter),
			DiameterOut: normalizeDiameter(t.DiameterOut),
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

	result.FilteredCount = filteredCount

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
			result.Error = err
			result.CompletedAt = time.Now()
			result.Duration = result.CompletedAt.Sub(result.StartedAt)
			return result, fmt.Errorf("failed to save tyres batch: %w", err)
		}

		totalNew += newCount
		totalSkipped += skippedCount

		s.logger.Info("Saved tyres batch",
			"batch", fmt.Sprintf("%d-%d", i, end),
			"new", newCount,
			"skipped_duplicates", skippedCount,
		)
	}

	result.NewCount = totalNew
	result.SkippedCount = totalSkipped

	// Get total count in database
	totalInDB, err := s.repo.CountTyres(ctx)
	if err != nil {
		s.logger.Error("Failed to count tyres", "error", err)
	} else {
		result.TotalInDB = totalInDB
	}

	// Update sync run
	if err := s.repo.UpdateSyncRun(ctx, runID, "completed", len(tyres), totalNew, totalSkipped, ""); err != nil {
		result.Error = err
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
		return result, fmt.Errorf("failed to update sync run: %w", err)
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	s.logger.Info("Tyres nomenclature sync completed",
		"total", len(tyres),
		"new", totalNew,
		"skipped_duplicates", totalSkipped,
	)

	return result, nil
}

// SyncRimsWithResult downloads and syncs rims nomenclature and returns detailed statistics
func (s *NomenclatureService) SyncRimsWithResult(ctx context.Context) (*SyncResult, error) {
	result := &SyncResult{
		Type:      "rims",
		StartedAt: time.Now(),
	}

	s.logger.Info("Starting rims nomenclature sync")

	// Create sync run
	runID, err := s.repo.CreateSyncRun(ctx, "rims")
	if err != nil {
		result.Error = err
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
		return result, fmt.Errorf("failed to create sync run: %w", err)
	}

	// Download XML
	s.logger.Info("Downloading rims XML", "url", rimsXMLURL)
	xmlData, err := s.downloadXML(ctx, rimsXMLURL)
	if err != nil {
		s.repo.UpdateSyncRun(ctx, runID, "failed", 0, 0, 0, err.Error())
		result.Error = err
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
		return result, fmt.Errorf("failed to download rims XML: %w", err)
	}

	// Parse XML
	s.logger.Info("Parsing rims XML", "size_bytes", len(xmlData))
	var root fourtochki.RimsXMLRoot
	if err := xml.Unmarshal(xmlData, &root); err != nil {
		s.repo.UpdateSyncRun(ctx, runID, "failed", 0, 0, 0, err.Error())
		result.Error = err
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
		return result, fmt.Errorf("failed to parse rims XML: %w", err)
	}

	s.logger.Info("Parsed rims", "count", len(root.Rims))

	// Convert to domain models
	rims := make([]domain.NomenclatureRim, 0, len(root.Rims))
	for _, r := range root.Rims {
		rims = append(rims, domain.NomenclatureRim{
			CAE:           r.CAE,
			Name:          r.Name,
			Width:         normalizeDiameter(r.Width),
			Diameter:      normalizeDiameter(r.Diameter),
			BoltsCount:    r.BoltsCount,
			BoltsSpacing:  normalizeDiameter(r.BoltsSpacing),
			BoltsSpacing2: normalizeDiameter(r.BoltsSpacing2),
			ET:            normalizeDiameter(r.ET),
			DIA:           normalizeDiameter(r.DIA),
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
			result.Error = err
			result.CompletedAt = time.Now()
			result.Duration = result.CompletedAt.Sub(result.StartedAt)
			return result, fmt.Errorf("failed to save rims batch: %w", err)
		}

		totalNew += newCount
		totalSkipped += skippedCount

		s.logger.Info("Saved rims batch",
			"batch", fmt.Sprintf("%d-%d", i, end),
			"new", newCount,
			"skipped_duplicates", skippedCount,
		)
	}

	result.NewCount = totalNew
	result.SkippedCount = totalSkipped

	// Get total count in database
	totalInDB, err := s.repo.CountRims(ctx)
	if err != nil {
		s.logger.Error("Failed to count rims", "error", err)
	} else {
		result.TotalInDB = totalInDB
	}

	// Update sync run
	if err := s.repo.UpdateSyncRun(ctx, runID, "completed", len(rims), totalNew, totalSkipped, ""); err != nil {
		result.Error = err
		result.CompletedAt = time.Now()
		result.Duration = result.CompletedAt.Sub(result.StartedAt)
		return result, fmt.Errorf("failed to update sync run: %w", err)
	}

	result.CompletedAt = time.Now()
	result.Duration = result.CompletedAt.Sub(result.StartedAt)

	s.logger.Info("Rims nomenclature sync completed",
		"total", len(rims),
		"new", totalNew,
		"skipped_duplicates", totalSkipped,
	)

	return result, nil
}
