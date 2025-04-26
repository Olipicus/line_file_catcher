package config

import (
	"fmt"
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

	// Google Drive configuration
	DriveEnabled     bool
	DriveCredentials string
	DriveTokenFile   string
	DriveFolder      string
	DriveRetryCount  int
}

// Load returns a Config struct populated with values from environment variables
func Load() *Config {
	// Load .env file if it exists
	godotenv.Load()

	config := &Config{
		ChannelSecret:    getEnv("LINE_CHANNEL_SECRET", ""),
		ChannelToken:     getEnv("LINE_CHANNEL_TOKEN", ""),
		Port:             getEnv("PORT", "8080"),
		StorageDir:       getEnv("STORAGE_DIR", "./storage"),
		LogDir:           getEnv("LOG_DIR", "./logs"),
		Debug:            getEnv("DEBUG", "false") == "true",
		DriveEnabled:     getEnv("DRIVE_ENABLED", "false") == "true",
		DriveCredentials: getEnv("DRIVE_CREDENTIALS", "./credentials.json"),
		DriveTokenFile:   getEnv("DRIVE_TOKEN_FILE", "./token.json"),
		DriveFolder:      getEnv("DRIVE_FOLDER", "LineFileCatcher"),
		DriveRetryCount:  getIntEnv("DRIVE_RETRY_COUNT", 3),
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

// getIntEnv retrieves an environment variable as integer or returns a default value
func getIntEnv(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue := defaultValue
	if _, err := fmt.Sscanf(value, "%d", &intValue); err != nil {
		log.Printf("Warning: Invalid value for %s, using default: %d", key, defaultValue)
		return defaultValue
	}

	return intValue
}

// GetMediaDir returns the path to the directory where media should be stored for a given date
func (c *Config) GetMediaDir(dateStr string) (string, error) {
	dir := filepath.Join(c.StorageDir, dateStr)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	return dir, nil
}
