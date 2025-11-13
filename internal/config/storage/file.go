package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

// FileStorage implements Storage interface for file-based storage.
// This is the authoritative source for configuration data.
type FileStorage struct {
	writer Writer
}

// NewFileStorage creates a new FileStorage instance.
func NewFileStorage() *FileStorage {
	return &FileStorage{
		writer: NewFileWriter(),
	}
}

// NewFileStorageWithWriter creates a new FileStorage with a custom writer (for testing).
func NewFileStorageWithWriter(writer Writer) *FileStorage {
	return &FileStorage{
		writer: writer,
	}
}

// Load reads configuration from a file.
func (f *FileStorage) Load(filePath string, target interface{}) error {
	// Ensure parent directory exists
	if err := EnsureParentDir(filePath); err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// File doesn't exist, return nil to indicate no data (caller should handle defaults)
		return nil
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Unmarshal JSON
	if len(data) == 0 {
		// Empty file, return nil (caller should handle defaults)
		return nil
	}

	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return nil
}

// Save writes configuration to a file atomically.
func (f *FileStorage) Save(filePath string, data interface{}) error {
	return SaveJSON(f.writer, data, filePath, 0o644, nil)
}

