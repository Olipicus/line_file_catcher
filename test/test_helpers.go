package test

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestHelpers contains utility functions for testing
type TestHelpers struct {
	t           *testing.T
	testDataDir string
}

// NewTestHelpers creates a new test helpers instance
func NewTestHelpers(t *testing.T) *TestHelpers {
	return &TestHelpers{
		t:           t,
		testDataDir: "../test_data",
	}
}

// GetTestFilePath returns the absolute path to a test file
func (h *TestHelpers) GetTestFilePath(filename string) string {
	return filepath.Join(h.testDataDir, filename)
}

// CopyTestFile copies a test file to a destination path
func (h *TestHelpers) CopyTestFile(sourcePath, destPath string) error {
	// Open source file
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// Copy contents
	_, err = io.Copy(destFile, sourceFile)
	return err
}

// ReadTestFile reads the contents of a test file
func (h *TestHelpers) ReadTestFile(filename string) ([]byte, error) {
	return os.ReadFile(h.GetTestFilePath(filename))
}

// AssertFileExistsWithPattern checks if a file exists in the directory with a specific pattern
func (h *TestHelpers) AssertFileExistsWithPattern(dir, pattern string) bool {
	files, err := os.ReadDir(dir)
	if err != nil {
		h.t.Errorf("Failed to read directory %s: %v", dir, err)
		return false
	}

	for _, file := range files {
		if strings.Contains(file.Name(), pattern) {
			return true
		}
	}

	h.t.Errorf("Expected to find file with pattern '%s' in directory %s, but none found", pattern, dir)
	return false
}

// CleanupTestDir removes all files in a directory but keeps the directory itself
func (h *TestHelpers) CleanupTestDir(dir string) error {
	d, err := os.Open(dir)
	if err != nil {
		return err
	}
	defer d.Close()

	names, err := d.Readdirnames(-1)
	if err != nil {
		return err
	}

	for _, name := range names {
		err = os.RemoveAll(filepath.Join(dir, name))
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateTestDir creates a test directory if it doesn't exist
func (h *TestHelpers) CreateTestDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
