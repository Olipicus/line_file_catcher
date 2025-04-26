package common

// CloudStorage defines the interface for cloud storage providers
type CloudStorage interface {
	// Initialize sets up the cloud storage service
	Initialize() error

	// UploadFile uploads a local file to cloud storage
	UploadFile(localPath, remoteFolder string) (string, error)

	// CreateFolder creates a folder in cloud storage if it doesn't exist
	CreateFolder(folderPath string) (string, error)

	// GetBackupStats returns statistics about the cloud storage usage
	GetBackupStats() map[string]interface{}
}
