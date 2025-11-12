package base

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureDir ensures the directory for the given file path exists
func EnsureDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}
	return nil
}
