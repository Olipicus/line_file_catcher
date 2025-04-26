package common

// CloudStorage defines the interface for cloud storage providers
type CloudStorage interface {
	// Initialize sets up the cloud storage service
	Initialize() error

	// UploadFile uploads a local file to cloud storage
	// Returns the file ID and error
	UploadFile(localPath, remoteFolder string) (string, error)

	// CreateFolder creates a folder in cloud storage if it doesn't exist
	CreateFolder(folderPath string) (string, error)

	// GetBackupStats returns statistics about the cloud storage usage
	GetBackupStats() map[string]interface{}

	// GetFileLink returns a shareable link for a file based on its ID
	GetFileLink(fileID string) (string, error)
}
