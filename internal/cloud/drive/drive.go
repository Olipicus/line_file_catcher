package drive

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"code.olipicus.com/line_file_catcher/internal/config"
	"code.olipicus.com/line_file_catcher/internal/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// DriveService implements CloudStorage interface for Google Drive
type DriveService struct {
	config      *config.Config
	logger      *utils.Logger
	service     *drive.Service
	folderCache map[string]string // Cache folder ID by path
	stats       DriveStats
	mu          sync.Mutex
}

// DriveStats stores statistics about Google Drive operations
type DriveStats struct {
	TotalUploaded      int64
	UploadCount        int
	FailedUploads      int
	RetryCount         int
	LastUploadTime     time.Time
	TotalUploadTime    time.Duration
	AverageUploadTime  time.Duration
	FolderCreatedCount int
}

// NewDriveService creates a new Google Drive service
func NewDriveService(cfg *config.Config, logger *utils.Logger) *DriveService {
	return &DriveService{
		config:      cfg,
		logger:      logger,
		folderCache: make(map[string]string),
		stats:       DriveStats{},
	}
}

// Initialize sets up the Google Drive service
func (d *DriveService) Initialize() error {
	d.logger.Info("Initializing Google Drive service")

	// Read the credentials file
	b, err := os.ReadFile(d.config.DriveCredentials)
	if err != nil {
		return fmt.Errorf("unable to read client secret file: %v", err)
	}

	// Parse the credentials
	config, err := google.ConfigFromJSON(b, drive.DriveFileScope)
	if err != nil {
		return fmt.Errorf("unable to parse client secret file: %v", err)
	}

	// Get or create token
	token, err := d.getToken(config)
	if err != nil {
		return fmt.Errorf("unable to get token: %v", err)
	}

	// Create the Drive client
	ctx := context.Background()
	client := config.Client(ctx, token)
	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("unable to create Drive service: %v", err)
	}

	d.service = srv
	d.logger.Info("Google Drive service initialized successfully")

	// Create the root folder if needed
	_, err = d.CreateFolder(d.config.DriveFolder)
	if err != nil {
		return fmt.Errorf("unable to create root folder: %v", err)
	}

	return nil
}

// getToken retrieves a token from a local file or requests a new one
func (d *DriveService) getToken(config *oauth2.Config) (*oauth2.Token, error) {
	tokenFile := d.config.DriveTokenFile

	// Check if token file exists
	token, err := d.tokenFromFile(tokenFile)
	if err == nil {
		return token, nil
	}

	// Token doesn't exist or is invalid, need to get a new one
	d.logger.Warning("No valid token found. Please authenticate via browser.")
	return nil, fmt.Errorf("no valid token found, please generate a token using the OAuth2 flow")
}

// tokenFromFile retrieves a token from a local file
func (d *DriveService) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	token := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(token)
	return token, err
}

// CreateFolder creates a folder in Google Drive if it doesn't exist
func (d *DriveService) CreateFolder(folderPath string) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Check cache first
	if id, ok := d.folderCache[folderPath]; ok {
		return id, nil
	}

	// Split path into components
	parts := strings.Split(strings.Trim(folderPath, "/"), "/")

	var parentID string = "root"
	var currentPath string

	// Create each folder in the path if it doesn't exist
	for _, part := range parts {
		if part == "" {
			continue
		}

		// Update current path
		if currentPath == "" {
			currentPath = part
		} else {
			currentPath = currentPath + "/" + part
		}

		// Check if this folder exists in cache
		if id, ok := d.folderCache[currentPath]; ok {
			parentID = id
			continue
		}

		// Search for the folder
		query := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder' and '%s' in parents and trashed=false", part, parentID)
		fileList, err := d.service.Files.List().Q(query).Fields("files(id, name)").Do()
		if err != nil {
			return "", fmt.Errorf("unable to search for folder %s: %v", part, err)
		}

		// Folder exists
		if len(fileList.Files) > 0 {
			folderID := fileList.Files[0].Id
			d.folderCache[currentPath] = folderID
			parentID = folderID
			continue
		}

		// Folder doesn't exist, create it
		folderMetadata := &drive.File{
			Name:     part,
			MimeType: "application/vnd.google-apps.folder",
			Parents:  []string{parentID},
		}

		folder, err := d.service.Files.Create(folderMetadata).Fields("id").Do()
		if err != nil {
			return "", fmt.Errorf("unable to create folder %s: %v", part, err)
		}

		d.folderCache[currentPath] = folder.Id
		parentID = folder.Id
		d.stats.FolderCreatedCount++
		d.logger.Debug("Created Google Drive folder: %s with ID: %s", part, folder.Id)
	}

	return parentID, nil
}

// UploadFile uploads a file to Google Drive
func (d *DriveService) UploadFile(localPath, remoteFolder string) (string, error) {
	// Start timing the upload
	startTime := time.Now()

	// Get the folder ID
	folderID, err := d.CreateFolder(remoteFolder)
	if err != nil {
		return "", fmt.Errorf("failed to create folder for upload: %v", err)
	}

	// Get file metadata
	filename := filepath.Base(localPath)

	// Create file metadata
	file := &drive.File{
		Name:    filename,
		Parents: []string{folderID},
	}

	// Open the local file
	content, err := os.Open(localPath)
	if err != nil {
		return "", fmt.Errorf("unable to open file for upload: %v", err)
	}
	defer content.Close()

	// Get file size for statistics
	fileInfo, err := content.Stat()
	if err != nil {
		return "", fmt.Errorf("unable to get file info: %v", err)
	}
	fileSize := fileInfo.Size()

	// Upload with retry logic
	var uploadedFile *drive.File
	var retryCount int

	for retryCount = 0; retryCount <= d.config.DriveRetryCount; retryCount++ {
		if retryCount > 0 {
			d.logger.Warning("Retrying upload for %s (attempt %d of %d)", filename, retryCount, d.config.DriveRetryCount)
			d.stats.RetryCount++

			// Reopen file for retry
			content.Close()
			content, err = os.Open(localPath)
			if err != nil {
				return "", fmt.Errorf("unable to reopen file for upload retry: %v", err)
			}

			// Wait before retry with exponential backoff
			time.Sleep(time.Duration(1<<retryCount) * time.Second)
		}

		// Create the file
		uploadedFile, err = d.service.Files.Create(file).Media(content).Fields("id, name, size").Do()
		if err == nil {
			break
		}

		// If we've reached the max retry count, fail
		if retryCount == d.config.DriveRetryCount {
			d.mu.Lock()
			d.stats.FailedUploads++
			d.mu.Unlock()
			return "", fmt.Errorf("failed to upload file after %d attempts: %v", retryCount+1, err)
		}
	}

	// Update statistics
	d.mu.Lock()
	d.stats.UploadCount++
	d.stats.TotalUploaded += fileSize
	d.stats.LastUploadTime = time.Now()

	uploadDuration := time.Since(startTime)
	d.stats.TotalUploadTime += uploadDuration
	d.stats.AverageUploadTime = d.stats.TotalUploadTime / time.Duration(d.stats.UploadCount)
	d.mu.Unlock()

	d.logger.Info("Successfully uploaded %s to Google Drive (ID: %s, Size: %d bytes) in %v",
		filename, uploadedFile.Id, fileSize, uploadDuration)

	return uploadedFile.Id, nil
}

// GetBackupStats returns the current backup statistics
func (d *DriveService) GetBackupStats() map[string]interface{} {
	d.mu.Lock()
	defer d.mu.Unlock()

	stats := map[string]interface{}{
		"totalUploaded":      d.stats.TotalUploaded,
		"uploadCount":        d.stats.UploadCount,
		"failedUploads":      d.stats.FailedUploads,
		"retryCount":         d.stats.RetryCount,
		"folderCreatedCount": d.stats.FolderCreatedCount,
		"averageUploadTime":  d.stats.AverageUploadTime.String(),
	}

	if !d.stats.LastUploadTime.IsZero() {
		stats["lastUploadTime"] = d.stats.LastUploadTime.Format(time.RFC3339)
	}

	return stats
}
