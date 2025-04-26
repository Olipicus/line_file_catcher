package config

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// LINE Bot API configuration
	ChannelSecret string
	ChannelToken  string

	// Server configuration
	Port string

	// Storage configuration
	StorageDir string

	// Logging configuration
	LogDir string
	Debug  bool
}

// Load returns a Config struct populated with values from environment variables
func Load() *Config {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		ChannelSecret: getEnv("LINE_CHANNEL_SECRET", ""),
		ChannelToken:  getEnv("LINE_CHANNEL_TOKEN", ""),
		Port:          getEnv("PORT", "8080"),
		StorageDir:    getEnv("STORAGE_DIR", "./storage"),
		LogDir:        getEnv("LOG_DIR", "./logs"),
		Debug:         getEnv("DEBUG", "false") == "true",
	}

	if config.ChannelSecret == "" || config.ChannelToken == "" {
		log.Fatal("LINE_CHANNEL_SECRET and LINE_CHANNEL_TOKEN must be set")
	}

	// Create storage directory if it doesn't exist
	if err := os.MkdirAll(config.StorageDir, 0755); err != nil {
		log.Fatalf("Failed to create storage directory: %v", err)
	}

	// Create log directory if it doesn't exist
	if err := os.MkdirAll(config.LogDir, 0755); err != nil {
		log.Fatalf("Failed to create log directory: %v", err)
	}

	return config
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetMediaDir returns the path to the directory where media should be stored for a given date
func (c *Config) GetMediaDir(dateStr string) (string, error) {
	dir := filepath.Join(c.StorageDir, dateStr)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return dir, nil
}
