package configutil

import (
	"encoding/json"
	"fmt"
	"os"
)

// SaveJSONConfig saves a config struct as JSON atomically using the provided storage
func SaveJSONConfig(storage Storage, config interface{}, filePath string, perm os.FileMode, ensureVersion func()) error {
	// Call ensureVersion if provided (to set version before marshaling)
	if ensureVersion != nil {
		ensureVersion()
	}

	// Marshal config with indentation
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write atomically using storage
	return storage.WriteFileAtomic(filePath, data, perm)
}
