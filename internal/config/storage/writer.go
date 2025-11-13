package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// Writer abstracts how configs hit the disk (handy for tests).
type Writer interface {
	WriteAtomic(filename string, data []byte, perm os.FileMode) error
}

// FileWriter is the default implementation used in production.
type FileWriter struct{}

// NewFileWriter returns a file-backed writer.
func NewFileWriter() *FileWriter {
	return &FileWriter{}
}

// WriteAtomic writes to a temporary file and then renames it to avoid corruption.
func (w *FileWriter) WriteAtomic(filename string, data []byte, perm os.FileMode) error {
	if err := EnsureParentDir(filename); err != nil {
		return err
	}

	dir := filepath.Dir(filename)
	tmp, err := os.CreateTemp(dir, filepath.Base(filename)+".tmp.*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	defer func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}()

	if _, err := tmp.Write(data); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}

	if err := tmp.Sync(); err != nil {
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	if err := tmp.Close(); err != nil {
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	if err := os.Chmod(tmpPath, perm); err != nil {
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}

	if err := os.Rename(tmpPath, filename); err != nil {
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	return nil
}
