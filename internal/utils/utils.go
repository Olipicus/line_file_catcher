package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"time"
)

// GenerateUniqueFilename creates a unique filename with the specified extension
// The format is: prefix_timestamp_randomString.extension
func GenerateUniqueFilename(prefix, extension string) (string, error) {
	// Get current timestamp
	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	// Generate random string (8 bytes = 16 hex chars)
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %v", err)
	}

	randomString := hex.EncodeToString(randomBytes)

	// Ensure extension starts with a dot
	if extension != "" && extension[0] != '.' {
		extension = "." + extension
	}

	// Create filename
	filename := fmt.Sprintf("%s_%d_%s%s", prefix, timestamp, randomString, extension)

	return filename, nil
}

// GetDateString returns the current date formatted as YYYY-MM-DD
func GetDateString() string {
	return time.Now().Format("2006-01-02")
}

// GetFileExtension extracts the extension from a filename
func GetFileExtension(filename string) string {
	return filepath.Ext(filename)
}

// GetContentType determines the file extension based on content type
func GetContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "video/mp4":
		return ".mp4"
	case "video/3gpp":
		return ".3gp"
	case "audio/mp4", "audio/mpeg", "audio/mp3":
		return ".mp3"
	default:
		return ".bin" // Default binary extension
	}
}
