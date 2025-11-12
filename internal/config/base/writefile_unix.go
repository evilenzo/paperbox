//go:build !windows
// +build !windows

package base

import (
	"os"

	"github.com/google/renameio/v2"
)

// WriteFileAtomic writes data to file atomically using renameio
func WriteFileAtomic(filename string, data []byte, perm os.FileMode) error {
	if err := EnsureDir(filename); err != nil {
		return err
	}
	return renameio.WriteFile(filename, data, perm, renameio.IgnoreUmask())
}

