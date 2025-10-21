package config

import (
	"os"
	"strconv"
)

type Config struct {
	Port        int
	Host        string
	DataDir     string
	LogLevel    string
	DatabaseURL string
	MasterKey   string
}

// New creates a new configuration from environment variables
func New() *Config {
	cfg := &Config{
		Port:        getEnvInt("CASSOCIAL_PORT", 8080),
		Host:        getEnv("CASSOCIAL_HOST", "0.0.0.0"),
		DataDir:     getEnv("CASSOCIAL_DATA", ""),
		LogLevel:    getEnv("CASSOCIAL_LOG_LEVEL", "info"),
		DatabaseURL: getEnv("CASSOCIAL_DATABASE_URL", ""),
		MasterKey:   getEnv("CASSOCIAL_MASTER_KEY", ""),
	}

	// Generate master key if not set
	if cfg.MasterKey == "" {
		cfg.MasterKey = generateMasterKey()
	}

	return cfg
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func generateMasterKey() string {
	// TODO: Generate a secure random master key
	// For now, return a placeholder
	return "change-me-in-production"
}
