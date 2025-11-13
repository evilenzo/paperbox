package configutil

import (
	"os"
)

// Storage defines the interface for configuration file operations
// This allows dependency injection for testing
type Storage interface {
	// WriteFileAtomic writes data to file atomically
	WriteFileAtomic(filename string, data []byte, perm os.FileMode) error
	// PatchConfig applies a partial update to a config struct
	PatchConfig(current interface{}, patch map[string]interface{}) (interface{}, error)
}

// FileStorage is the default implementation of Storage interface
type FileStorage struct{}

// NewFileStorage creates a new FileStorage instance
func NewFileStorage() *FileStorage {
	return &FileStorage{}
}

// WriteFileAtomic writes data to file atomically
func (s *FileStorage) WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	return WriteFileAtomic(filename, data, perm)
}

// PatchConfig applies a partial update to a config struct
func (s *FileStorage) PatchConfig(current interface{}, patch map[string]interface{}) (interface{}, error) {
	return PatchConfig(current, patch)
}
