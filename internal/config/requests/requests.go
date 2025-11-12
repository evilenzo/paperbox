package requests

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/adrg/xdg"
	"github.com/go-playground/validator/v10"
)

const (
	// CurrentVersion is the current version of the requests config format
	CurrentVersion = 1
	// RequestsFileName is the name of the requests config file
	RequestsFileName = "requests.json"
)

var (
	appDataDir   = path.Join(xdg.DataHome, "paperbox")
	requestsFile = path.Join(appDataDir, RequestsFileName)
	validate     *validator.Validate
)

func init() {
	validate = validator.New()

	// Register custom validators
	if err := validate.RegisterValidation("http_method", validateHTTPMethod); err != nil {
		panic(fmt.Sprintf("failed to register http_method validator: %v", err))
	}
}

// ItemType represents the type of an item
type ItemType string

const (
	ItemTypeRequest ItemType = "request"
	ItemTypeFolder  ItemType = "folder"
)

// Item represents a request or folder item
type Item struct {
	Type     ItemType `json:"type" validate:"required,oneof=request folder"`
	Name     string   `json:"name" validate:"required,min=1"`
	Method   string   `json:"method,omitempty" validate:"omitempty,http_method"`
	Path     string   `json:"path,omitempty" validate:"omitempty,min=1"`
	Children []string `json:"children,omitempty" validate:"omitempty,dive,required"`
}

// RequestsConfig represents the requests configuration
type RequestsConfig struct {
	Version int             `json:"version" validate:"required,min=1"`
	Values  map[string]Item `json:"values" validate:"required,dive,keys,required,endkeys"`
}

// NewRequestsConfig creates a new empty requests config
func NewRequestsConfig() *RequestsConfig {
	return &RequestsConfig{
		Version: CurrentVersion,
		Values:  make(map[string]Item),
	}
}

// Load loads the requests configuration from file
func Load() (*RequestsConfig, error) {
	// Create app data directory if it doesn't exist
	if _, err := os.Stat(appDataDir); os.IsNotExist(err) {
		err := os.MkdirAll(appDataDir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create app data directory: %w", err)
		}
	}

	// Create requests file if it doesn't exist
	if _, err := os.Stat(requestsFile); os.IsNotExist(err) {
		config := NewRequestsConfig()
		if err := Save(config); err != nil {
			return nil, fmt.Errorf("failed to create requests file: %w", err)
		}
		return config, nil
	}

	// Read requests file
	data, err := os.ReadFile(requestsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read requests file: %w", err)
	}

	// Parse config
	var config RequestsConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse requests file: %w", err)
	}

	// Migrate config if needed
	if err := migrateConfig(&config); err != nil {
		return nil, fmt.Errorf("failed to migrate requests config: %w", err)
	}

	// Validate config
	if err := Validate(&config); err != nil {
		return nil, fmt.Errorf("requests config validation failed: %w", err)
	}

	return &config, nil
}

// Save saves the requests configuration to file
func Save(config *RequestsConfig) error {
	// Ensure version is set
	if config.Version == 0 {
		config.Version = CurrentVersion
	}

	// Create app data directory if it doesn't exist
	if _, err := os.Stat(appDataDir); os.IsNotExist(err) {
		err := os.MkdirAll(appDataDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create app data directory: %w", err)
		}
	}

	// Marshal config
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal requests config: %w", err)
	}

	// Write config file
	if err := os.WriteFile(requestsFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write requests file: %w", err)
	}

	return nil
}

// migrateConfig migrates config from older versions to current version
func migrateConfig(config *RequestsConfig) error {
	if config.Version == 0 {
		// Version 0: old format without version field
		// Assume it's version 1 format
		config.Version = 1
	}

	if config.Version < CurrentVersion {
		// Perform migrations
		for version := config.Version; version < CurrentVersion; version++ {
			if err := migrateFromVersion(config, version); err != nil {
				return fmt.Errorf("failed to migrate from version %d: %w", version, err)
			}
		}
		config.Version = CurrentVersion
		// Save migrated config
		_ = Save(config) // Ignore errors, continue with default config
	}

	return nil
}

// migrateFromVersion migrates config from a specific version
func migrateFromVersion(config *RequestsConfig, fromVersion int) error {
	switch fromVersion {
	case 0:
		// Migration from version 0 to 1
		// No changes needed, just version field addition
		return nil
	default:
		return fmt.Errorf("unknown migration from version %d", fromVersion)
	}
}

// MarshalRequests marshals a map of items to JSON (for Requests structure)
func MarshalRequests(values map[string]Item) ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"values": values,
	})
}

// UnmarshalRequests unmarshals JSON to a map of items (for Requests structure)
func UnmarshalRequests(data []byte) (map[string]Item, error) {
	var aux struct {
		Values map[string]Item `json:"values"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return nil, err
	}
	if aux.Values == nil {
		aux.Values = make(map[string]Item)
	}
	return aux.Values, nil
}
