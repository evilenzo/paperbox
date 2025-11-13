package storage

import (
	"encoding/json"
	"fmt"
)

// MergePatch copies the current config, applies the patch, and decodes into target.
func MergePatch(current interface{}, patch map[string]interface{}, target interface{}) error {
	currentJSON, err := json.Marshal(current)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal(currentJSON, &configMap); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}

	for key, value := range patch {
		configMap[key] = value
	}

	mergedJSON, err := json.Marshal(configMap)
	if err != nil {
		return fmt.Errorf("failed to marshal merged config: %w", err)
	}

	if err := json.Unmarshal(mergedJSON, target); err != nil {
		return fmt.Errorf("failed to unmarshal merged config: %w", err)
	}

	return nil
}
