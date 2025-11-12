package base

import (
	"encoding/json"
	"fmt"
)

// PatchConfig applies a partial update to a config struct using JSON marshal/unmarshal
// This is a helper function that can be used by any config manager
func PatchConfig(current interface{}, patch map[string]interface{}) (interface{}, error) {
	// Convert current config to map for merging
	configJSON, err := json.Marshal(current)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	var configMap map[string]interface{}
	if err := json.Unmarshal(configJSON, &configMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Merge patch into config map
	for key, value := range patch {
		configMap[key] = value
	}

	return configMap, nil
}

// UnmarshalPatchedConfig unmarshals a patched config map back to the target struct
func UnmarshalPatchedConfig(patchedMap map[string]interface{}, target interface{}) error {
	mergedJSON, err := json.Marshal(patchedMap)
	if err != nil {
		return fmt.Errorf("failed to marshal merged config: %w", err)
	}

	if err := json.Unmarshal(mergedJSON, target); err != nil {
		return fmt.Errorf("failed to unmarshal merged config: %w", err)
	}

	return nil
}
