package user

import (
	"context"
	"fmt"
	"os"
	"path"

	"paperbox/internal/config/core"
	"paperbox/internal/config/storage"

	"github.com/adrg/xdg"
	"github.com/wailsapp/wails/v2/pkg/logger"
)

const (
	// CurrentVersion is the current version of the user config format
	CurrentVersion = 1
	// ConfigFileName is the name of the user config file
	ConfigFileName = "config.json"
)

var (
	appDataDir = path.Join(xdg.DataHome, "paperbox")
	configFile = path.Join(appDataDir, ConfigFileName)
)

// Config represents the user configuration
type Config struct {
	Version  int    `json:"version"`
	Theme    string `json:"theme"`    // "light" | "dark" | "auto"
	FontSize int    `json:"fontSize"` // Font size in pixels
	BaseURL  string `json:"baseURL"`  // Base URL for API requests
}

// DefaultConfig returns a new config with default values
func DefaultConfig() *Config {
	return &Config{
		Version:  CurrentVersion,
		Theme:    "light",
		FontSize: 14,
		BaseURL:  "",
	}
}

// Manager manages the user configuration
type Manager struct {
	*core.BaseManager[Config]
}

// loadUserConfig loads user config from file, creating default if file doesn't exist
func loadUserConfig() (*Config, error) {
	// Ensure directory exists
	if err := storage.EnsureParentDir(configFile); err != nil {
		return nil, fmt.Errorf("failed to ensure parent directory: %w", err)
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return DefaultConfig(), nil
	}

	// Load from file using FileStorage
	fileStorage := storage.NewFileStorage()
	var cfg Config
	if err := fileStorage.Load(configFile, &cfg); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Ensure version is set
	if cfg.Version == 0 {
		cfg.Version = CurrentVersion
	}

	return &cfg, nil
}

// NewManager creates a new config manager
func NewManager(storage storage.Storage) *Manager {
	return &Manager{
		BaseManager: core.NewBaseManager(core.BaseManagerOptions[Config]{
			Storage:    storage,
			ConfigFile: configFile,
			EventName:  "config",
			Loader:     loadUserConfig,
			Validator:  nil, // No validation for user config
			EnsureFunc: func(cfg *Config) {
				if cfg.Version == 0 {
					cfg.Version = CurrentVersion
				}
			},
		}),
	}
}

// NewManagerWithWriter creates a new config manager with custom writer (for testing)
func NewManagerWithWriter(writer storage.Writer) *Manager {
	fileStorage := storage.NewFileStorageWithWriter(writer)
	coordinator := storage.NewStorageCoordinator(fileStorage, nil, nil)

	return &Manager{
		BaseManager: core.NewBaseManager(core.BaseManagerOptions[Config]{
			Storage:    coordinator,
			ConfigFile: configFile,
			EventName:  "config",
			Loader:     loadUserConfig,
			Validator:  nil, // No validation for user config
			EnsureFunc: func(cfg *Config) {
				if cfg.Version == 0 {
					cfg.Version = CurrentVersion
				}
			},
		}),
	}
}

// SetContext sets the Wails runtime context for emitting events
func (m *Manager) SetContext(ctx context.Context, log logger.Logger) {
	m.BaseManager.SetContext(ctx, log)
}

// Get returns a copy of the current configuration (implements ManagerInterface)
func (m *Manager) Get() interface{} {
	return m.GetConfig()
}

// GetConfig returns the user config (type-safe version)
func (m *Manager) GetConfig() *Config {
	return m.BaseManager.Get()
}

// Patch applies a partial update to the configuration
func (m *Manager) Patch(patch map[string]interface{}) error {
	return m.BaseManager.Patch(patch)
}
