package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration
type Config struct {
	// General
	AppEnv   string
	LogLevel string

	// HTTP
	HTTPPort string

	// Database
	DatabaseDSN string

	// Fourtochki (4tochki) Provider
	FourtochkiWSDLURL    string
	FourtochkiLogin      string
	FourtochkiPassword   string
	FourtochkiBatchSize  int
	FourtochkiTimeout    time.Duration
	FourtochkiRetryCount int
	FourtochkiRetryDelay time.Duration
	FourtochkiBatchDelay time.Duration

	// Severavto Provider
	SeveravtoBaseURL string
	SeveravtoAPIKey  string
	SeveravtoTimeout string

	// Email
	EmailEnabled        bool
	EmailSMTPHost       string
	EmailSMTPPort       int
	EmailUsername       string
	EmailPassword       string
	EmailFrom           string
	EmailNotificationTo string
}


// Load reads configuration from environment variables
func Load() *Config {
	return &Config{
		// General
		AppEnv:   getEnv("APP_ENV", "development"),
		LogLevel: getEnv("LOG_LEVEL", "info"),

		// HTTP
		HTTPPort: getEnv("HTTP_PORT", "8080"),

		// Database
		DatabaseDSN: getEnv("DATABASE_DSN", getEnv("DB_DSN", "")),

		// Fourtochki
		FourtochkiWSDLURL:    getEnv("FOURTOCHKI_WSDL_URL", "http://api-b2b.4tochki.ru/WCF/ClientService.svc?wsdl"),
		FourtochkiLogin:      getEnv("FOURTOCHKI_LOGIN", ""),
		FourtochkiPassword:   getEnv("FOURTOCHKI_PASSWORD", ""),
		FourtochkiBatchSize:  getEnvAsInt("FOURTOCHKI_BATCH_SIZE", 100),
		FourtochkiTimeout:    getEnvAsDuration("FOURTOCHKI_TIMEOUT", 30*time.Second),
		FourtochkiRetryCount: getEnvAsInt("FOURTOCHKI_RETRY_COUNT", 3),
		FourtochkiRetryDelay: getEnvAsDuration("FOURTOCHKI_RETRY_DELAY", 5*time.Second),
		FourtochkiBatchDelay: getEnvAsDuration("FOURTOCHKI_BATCH_DELAY", 1*time.Second),

		// Severavto
		SeveravtoBaseURL: getEnv("SEVERAVTO_BASE_URL", "https://webmim.svrauto.ru"),
		SeveravtoAPIKey:  getEnv("SEVERAVTO_API_KEY", ""),
		SeveravtoTimeout: getEnv("SEVERAVTO_TIMEOUT", "60s"),

		// Email
		EmailEnabled:        getEnvAsBool("EMAIL_ENABLED", true),
		EmailSMTPHost:       getEnv("EMAIL_SMTP_HOST", "mail.hosting.reg.ru"),
		EmailSMTPPort:       getEnvAsInt("EMAIL_SMTP_PORT", 465),
		EmailUsername:       getEnv("EMAIL_USERNAME", ""),
		EmailPassword:       getEnv("EMAIL_PASSWORD", ""),
		EmailFrom:           getEnv("EMAIL_FROM", "admin@etalon-shina.ru"),
		EmailNotificationTo: getEnv("EMAIL_NOTIFICATION_TO", "v.boyarkin@etalon-shina.ru"),
	}
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
