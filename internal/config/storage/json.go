package storage

import (
	"encoding/json"
	"fmt"
	"os"
)

// SaveJSON marshals cfg with indentation and writes it atomically.
func SaveJSON(writer Writer, cfg interface{}, filePath string, perm os.FileMode, ensure func()) error {
	if ensure != nil {
		ensure()
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return writer.WriteAtomic(filePath, data, perm)
}
