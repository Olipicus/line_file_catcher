package media

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"code.olipicus.com/line_file_catcher/internal/cloud/common"
	"code.olipicus.com/line_file_catcher/internal/cloud/drive"
	"code.olipicus.com/line_file_catcher/internal/config"
	"code.olipicus.com/line_file_catcher/internal/utils"
	"github.com/line/line-bot-sdk-go/v7/linebot"
)

// Stats tracks file processing statistics
type Stats struct {
	ImageCount int       `json:"imageCount"`
	VideoCount int       `json:"videoCount"`
	AudioCount int       `json:"audioCount"`
	FileCount  int       `json:"fileCount"`
	TotalBytes int64     `json:"totalBytes"`
	StartTime  time.Time `json:"startTime"`
	mu         sync.Mutex
}

// MediaStore handles the downloading and storing of media files
type MediaStore struct {
	config     *config.Config
	logger     *utils.Logger
	cloudStore common.CloudStorage
	downloadWg sync.WaitGroup
	uploadWg   sync.WaitGroup
	stats      Stats
}

// NewMediaStore creates a new MediaStore instance
func NewMediaStore(cfg *config.Config, logger *utils.Logger) *MediaStore {
	ms := &MediaStore{
		config: cfg,
		logger: logger,
		stats: Stats{
			StartTime: time.Now(),
		},
	}

	// Initialize cloud storage if enabled
	if cfg.DriveEnabled {
		driveService := drive.NewDriveService(cfg, logger)
		err := driveService.Initialize()
		if err != nil {
			logger.Error("Failed to initialize Google Drive: %v", err)
			logger.Warning("Google Drive backup will be disabled")
		} else {
			ms.cloudStore = driveService
			logger.Info("Google Drive backup enabled")
		}
	} else {
		logger.Info("Google Drive backup disabled")
	}

	return ms
}

// SaveMedia saves media content from a LINE MessageContentResponse
func (ms *MediaStore) SaveMedia(messageID, messageType string, content *linebot.MessageContentResponse) (string, error) {
	// Use current date for organizing files
	dateStr := utils.GetDateString()

	ms.logger.Debug("Saving %s media with ID %s", messageType, messageID)

	// Get directory for storing files based on date
	storageDir, err := ms.config.GetMediaDir(dateStr)
	if err != nil {
		return "", fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Determine file extension based on content type
	contentType := content.ContentType
	ms.logger.Debug("Media %s has content type: %s", messageID, contentType)
	extension := utils.GetContentType(contentType)

	// Generate a unique filename
	filename, err := utils.GenerateUniqueFilename(messageType, extension)
	if err != nil {
		return "", fmt.Errorf("failed to generate filename: %v", err)
	}

	// Full path to save the file
	filePath := filepath.Join(storageDir, filename)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Copy content to file
	bytesWritten, err := io.Copy(file, content.Content)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// Update statistics
	ms.updateStats(messageType, bytesWritten)

	ms.logger.Info("Saved %s media file of %d bytes to %s", messageType, bytesWritten, filePath)

	// Upload to cloud storage if enabled
	ms.uploadToCloudAsync(filePath, dateStr)

	return filePath, nil
}

// uploadToCloudAsync uploads a file to cloud storage asynchronously
func (ms *MediaStore) uploadToCloudAsync(filePath, folderPath string) {
	// Skip if cloud storage is not configured
	if ms.cloudStore == nil {
		return
	}

	ms.uploadWg.Add(1)
	go func() {
		defer ms.uploadWg.Done()

		ms.logger.Debug("Starting cloud upload for %s to folder %s", filePath, folderPath)

		// Build the remote folder path using the cloud provider's base folder and the date subfolder
		remoteFolder := filepath.Join(ms.config.DriveFolder, folderPath)

		// Upload the file
		fileID, err := ms.cloudStore.UploadFile(filePath, remoteFolder)
		if err != nil {
			ms.logger.Error("Failed to upload file to cloud storage: %v", err)
			return
		}

		ms.logger.Info("Successfully uploaded %s to cloud storage (ID: %s)", filePath, fileID)
	}()
}

// updateStats updates the statistics counter safely
func (ms *MediaStore) updateStats(mediaType string, bytes int64) {
	ms.stats.mu.Lock()
	defer ms.stats.mu.Unlock()

	ms.stats.TotalBytes += bytes

	switch mediaType {
	case "image":
		ms.stats.ImageCount++
	case "video":
		ms.stats.VideoCount++
	case "audio":
		ms.stats.AudioCount++
	case "file":
		ms.stats.FileCount++
	}
}

// GetStats returns a copy of the current statistics
func (ms *MediaStore) GetStats() Stats {
	ms.stats.mu.Lock()
	defer ms.stats.mu.Unlock()

	// Return a copy to avoid race conditions
	return Stats{
		ImageCount: ms.stats.ImageCount,
		VideoCount: ms.stats.VideoCount,
		AudioCount: ms.stats.AudioCount,
		FileCount:  ms.stats.FileCount,
		TotalBytes: ms.stats.TotalBytes,
		StartTime:  ms.stats.StartTime,
	}
}

// GetCloudStats returns statistics about cloud storage if available
func (ms *MediaStore) GetCloudStats() map[string]interface{} {
	if ms.cloudStore == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	stats := ms.cloudStore.GetBackupStats()
	stats["enabled"] = true

	return stats
}

// DownloadMedia downloads media from a URL and saves it to disk
func (ms *MediaStore) DownloadMedia(messageID, messageType string, contentURL string, headers map[string]string) (string, error) {
	// Use current date for organizing files
	dateStr := utils.GetDateString()

	ms.logger.Debug("Downloading %s media with ID %s", messageType, messageID)

	// Get directory for storing files based on date
	storageDir, err := ms.config.GetMediaDir(dateStr)
	if err != nil {
		return "", fmt.Errorf("failed to create storage directory: %v", err)
	}

	// Create request to download the content
	req, err := http.NewRequest("GET", contentURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	// Add required headers (e.g., Authorization)
	for key, value := range headers {
		req.Header.Add(key, value)
	}

	// Execute the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download media: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to download media, status code: %d", resp.StatusCode)
	}

	// Determine file extension based on content type
	contentType := resp.Header.Get("Content-Type")
	ms.logger.Debug("Media %s has content type: %s", messageID, contentType)
	extension := utils.GetContentType(contentType)

	// Generate a unique filename
	filename, err := utils.GenerateUniqueFilename(messageType, extension)
	if err != nil {
		return "", fmt.Errorf("failed to generate filename: %v", err)
	}

	// Full path to save the file
	filePath := filepath.Join(storageDir, filename)

	// Create the file
	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Copy content to file
	bytesWritten, err := io.Copy(file, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// Update statistics
	ms.updateStats(messageType, bytesWritten)

	ms.logger.Info("Saved %s media file of %d bytes to %s", messageType, bytesWritten, filePath)

	// Upload to cloud storage if enabled
	ms.uploadToCloudAsync(filePath, dateStr)

	return filePath, nil
}

// AddToDownloadQueue adds a media download task to the queue
func (ms *MediaStore) AddToDownloadQueue(messageID, messageType string, contentURL string, headers map[string]string) {
	ms.downloadWg.Add(1)

	ms.logger.Info("Queuing download for %s media with ID %s", messageType, messageID)

	go func() {
		defer ms.downloadWg.Done()

		filePath, err := ms.DownloadMedia(messageID, messageType, contentURL, headers)
		if err != nil {
			ms.logger.Error("Error downloading media %s: %v", messageID, err)
			return
		}

		ms.logger.Info("Successfully downloaded and saved media %s to %s", messageID, filePath)
	}()
}

// WaitForDownloads waits for all queued downloads to complete
func (ms *MediaStore) WaitForDownloads() {
	ms.logger.Info("Waiting for pending downloads to complete...")
	ms.downloadWg.Wait()
	ms.logger.Info("All downloads completed")
}

// WaitForUploads waits for all cloud uploads to complete
func (ms *MediaStore) WaitForUploads() {
	if ms.cloudStore == nil {
		return
	}

	ms.logger.Info("Waiting for pending cloud uploads to complete...")
	ms.uploadWg.Wait()
	ms.logger.Info("All cloud uploads completed")
}

// WaitForAll waits for all pending downloads and uploads to complete
func (ms *MediaStore) WaitForAll() {
	ms.WaitForDownloads()
	ms.WaitForUploads()
}
