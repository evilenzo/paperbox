package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// EnsureParentDir creates the parent directory for a file if it does not exist.
func EnsureParentDir(filePath string) error {
	dir := filepath.Dir(filePath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to stat directory: %w", err)
	}
	return nil
}
