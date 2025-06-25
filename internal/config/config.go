package config

import (
	"os"
	"strconv"
)

// Config holds all application configuration
type Config struct {
	Port         string
	DatabasePath string
	LogFilePath  string
}

// Load creates a new configuration from environment variables with defaults
func Load() *Config {
	return &Config{
		Port:         getEnv("PORT", "9090"),
		DatabasePath: getEnv("DB_PATH", "bookmarks.db"),
		LogFilePath:  getEnv("LOG_FILE", "bookminderapi.log"),
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetPortInt returns the port as an integer
func (c *Config) GetPortInt() int {
	port, err := strconv.Atoi(c.Port)
	if err != nil {
		return 9090 // default fallback
	}
	return port
}