package httpserver

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prokoleso/etalon-price-api/internal/providers"
	"github.com/prokoleso/etalon-price-api/internal/repository/postgres"
	"github.com/prokoleso/etalon-price-api/internal/service"
)

// Server represents HTTP server
type Server struct {
	httpServer  *http.Server
	syncService *service.SyncService
	db          *postgres.DB
	registry    *providers.Registry
	logger      *slog.Logger
}

// Config holds server configuration
type Config struct {
	Port        string
	SyncService *service.SyncService
	DB          *postgres.DB
	Registry    *providers.Registry
	Logger      *slog.Logger
}

// New creates a new HTTP server
func New(cfg Config) *Server {
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	s := &Server{
		syncService: cfg.SyncService,
		db:          cfg.DB,
		registry:    cfg.Registry,
		logger:      cfg.Logger,
	}

	mux := http.NewServeMux()
	s.registerRoutes(mux)

	s.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      s.loggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s
}

// registerRoutes registers all HTTP routes
func (s *Server) registerRoutes(mux *http.ServeMux) {
	// Health checks
	mux.HandleFunc("/healthz", s.handleHealth)
	mux.HandleFunc("/readyz", s.handleReady)
	mux.HandleFunc("/version", s.handleVersion)

	// API endpoints
	mux.HandleFunc("/sync/4tochki", s.handleSync4Tochki)
	mux.HandleFunc("/stats", s.handleStats)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	s.logger.Info("HTTP server starting", "addr", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("HTTP server shutting down...")
	return s.httpServer.Shutdown(ctx)
}

// loggingMiddleware logs all HTTP requests
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Create response wrapper to capture status code
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		duration := time.Since(start)

		s.logger.Info("HTTP request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration", duration,
			"remote_addr", r.RemoteAddr,
		)
	})
}

// responseWriter wraps http.ResponseWriter to capture status code
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// writeJSON writes JSON response
func (s *Server) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.logger.Error("Failed to encode JSON response", "error", err)
	}
}

// writeError writes error response
func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	s.writeJSON(w, status, map[string]string{
		"error": message,
	})
}
