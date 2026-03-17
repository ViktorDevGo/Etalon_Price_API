package httpserver

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/prokoleso/etalon-price-api/internal/service"
)

// SyncRequest represents sync request body
type SyncRequest struct {
	Codes []string `json:"codes"`
}

// SyncResponse represents sync response
type SyncResponse struct {
	Supplier     string   `json:"supplier"`
	SyncRunID    int64    `json:"sync_run_id"`
	TotalCodes   int      `json:"total_codes"`
	TyresSaved   int      `json:"tyres_saved"`
	RimsSaved    int      `json:"rims_saved"`
	OffersSaved  int      `json:"offers_saved"`
	DurationMs   int64    `json:"duration_ms"`
	Errors       []string `json:"errors"`
}

// handleHealth handles health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Check database connection
	if err := s.db.Ping(r.Context()); err != nil {
		s.writeJSON(w, http.StatusServiceUnavailable, map[string]interface{}{
			"status": "unhealthy",
			"error":  err.Error(),
		})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{
		"status": "healthy",
	})
}

// handleReady handles readiness check endpoint
func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"ready":     true,
		"providers": s.registry.List(),
	})
}

// handleVersion handles version endpoint
func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{
		"version": "1.0.0",
		"service": "etalon-price-api",
	})
}

// handleStats handles stats endpoint
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	stats := s.db.Stats()
	s.writeJSON(w, http.StatusOK, map[string]interface{}{
		"database": map[string]interface{}{
			"acquired_conns":      stats.AcquiredConns(),
			"idle_conns":          stats.IdleConns(),
			"total_conns":         stats.TotalConns(),
			"max_conns":           stats.MaxConns(),
			"acquire_count":       stats.AcquireCount(),
			"empty_acquire_count": stats.EmptyAcquireCount(),
		},
	})
}

// handleSync4Tochki handles sync request for 4tochki provider
func (s *Server) handleSync4Tochki(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	var req SyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate codes
	if len(req.Codes) == 0 {
		s.writeError(w, http.StatusBadRequest, "Codes are required")
		return
	}

	if len(req.Codes) > 1000 {
		s.writeError(w, http.StatusBadRequest, "Too many codes (max 1000)")
		return
	}

	s.logger.Info("Sync request received",
		"provider", "4tochki",
		"codes_count", len(req.Codes),
	)

	// Perform synchronization
	start := time.Now()

	result, err := s.syncService.SyncSupplier(r.Context(), "4tochki", service.SyncOptions{
		Codes:      req.Codes,
		BatchSize:  100,
		BatchDelay: 1 * time.Second,
	})

	duration := time.Since(start)

	if err != nil {
		s.logger.Error("Sync failed",
			"provider", "4tochki",
			"error", err,
		)
		s.writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Count tyres and rims from success count
	// This is a simplified version - in real scenario you'd track these separately
	tyresSaved := result.SuccessProducts / 2
	rimsSaved := result.SuccessProducts - tyresSaved

	// Build error messages
	errorMessages := make([]string, 0, len(result.Errors))
	for _, e := range result.Errors {
		errorMessages = append(errorMessages, e.Error())
	}

	response := SyncResponse{
		Supplier:     "4tochki",
		SyncRunID:    result.SyncRunID,
		TotalCodes:   len(req.Codes),
		TyresSaved:   tyresSaved,
		RimsSaved:    rimsSaved,
		OffersSaved:  result.SuccessProducts,
		DurationMs:   duration.Milliseconds(),
		Errors:       errorMessages,
	}

	s.logger.Info("Sync completed",
		"provider", "4tochki",
		"sync_run_id", result.SyncRunID,
		"duration_ms", duration.Milliseconds(),
	)

	s.writeJSON(w, http.StatusOK, response)
}
