package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	App        AppConfig
	HTTP       HTTPConfig
	Database   DatabaseConfig
	Fourtochki FourtochkiConfig
	Email      EmailConfig
}

// AppConfig contains general application settings
type AppConfig struct {
	Env      string // development, production
	LogLevel string // debug, info, warn, error
}

// HTTPConfig contains HTTP server settings
type HTTPConfig struct {
	Port string
}

// DatabaseConfig contains database connection settings
type DatabaseConfig struct {
	DSN string
}

// FourtochkiConfig contains 4tochki provider settings
type FourtochkiConfig struct {
	WSDLURL    string
	Login      string
	Password   string
	BatchSize  int
	Timeout    time.Duration
	RetryCount int
	RetryDelay time.Duration
	BatchDelay time.Duration
}

// EmailConfig contains email notification settings
type EmailConfig struct {
	Enabled        bool
	SMTPHost       string
	SMTPPort       int
	Username       string
	Password       string
	From           string
	NotificationTo string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{
		App: AppConfig{
			Env:      getEnv("APP_ENV", "development"),
			LogLevel: getEnv("LOG_LEVEL", "info"),
		},
		HTTP: HTTPConfig{
			Port: getEnv("HTTP_PORT", "8080"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DB_DSN", ""),
		},
		Fourtochki: FourtochkiConfig{
			WSDLURL:    getEnv("FOURTOCHKI_WSDL_URL", "http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl"),
			Login:      getEnv("FOURTOCHKI_LOGIN", ""),
			Password:   getEnv("FOURTOCHKI_PASSWORD", ""),
			BatchSize:  getEnvAsInt("FOURTOCHKI_BATCH_SIZE", 100),
			Timeout:    getEnvAsDuration("FOURTOCHKI_TIMEOUT", 30*time.Second),
			RetryCount: getEnvAsInt("FOURTOCHKI_RETRY_COUNT", 3),
			RetryDelay: getEnvAsDuration("FOURTOCHKI_RETRY_DELAY", 5*time.Second),
			BatchDelay: getEnvAsDuration("FOURTOCHKI_BATCH_DELAY", 1*time.Second),
		},
		Email: EmailConfig{
			Enabled:        getEnvAsBool("EMAIL_ENABLED", true),
			SMTPHost:       getEnv("EMAIL_SMTP_HOST", "mail.hosting.reg.ru"),
			SMTPPort:       getEnvAsInt("EMAIL_SMTP_PORT", 465),
			Username:       getEnv("EMAIL_USERNAME", ""),
			Password:       getEnv("EMAIL_PASSWORD", ""),
			From:           getEnv("EMAIL_FROM", "admin@etalon-shina.ru"),
			NotificationTo: getEnv("EMAIL_NOTIFICATION_TO", "v.boyarkin@etalon-shina.ru"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate checks if all required configuration parameters are set
func (c *Config) Validate() error {
	if c.Database.DSN == "" {
		return fmt.Errorf("DB_DSN is required")
	}

	if c.Fourtochki.Login == "" {
		return fmt.Errorf("FOURTOCHKI_LOGIN is required")
	}

	if c.Fourtochki.Password == "" {
		return fmt.Errorf("FOURTOCHKI_PASSWORD is required")
	}

	if c.Fourtochki.BatchSize <= 0 {
		return fmt.Errorf("FOURTOCHKI_BATCH_SIZE must be positive")
	}

	if c.Fourtochki.Timeout <= 0 {
		return fmt.Errorf("FOURTOCHKI_TIMEOUT must be positive")
	}

	if c.Fourtochki.RetryCount < 0 {
		return fmt.Errorf("FOURTOCHKI_RETRY_COUNT must be non-negative")
	}

	validLogLevels := map[string]bool{"debug": true, "info": true, "warn": true, "error": true}
	if !validLogLevels[c.App.LogLevel] {
		return fmt.Errorf("LOG_LEVEL must be one of: debug, info, warn, error")
	}

	validEnvs := map[string]bool{"development": true, "production": true, "staging": true}
	if !validEnvs[c.App.Env] {
		return fmt.Errorf("APP_ENV must be one of: development, production, staging")
	}

	return nil
}

// getEnv reads an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt reads an environment variable as integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsDuration reads an environment variable as duration or returns a default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := time.ParseDuration(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvAsBool reads an environment variable as boolean or returns a default value
func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
