package user

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"

	"paperbox/internal/config/base"

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
	*base.BaseManager
	config *Config
}

// NewManager creates a new config manager
func NewManager() *Manager {
	return &Manager{
		BaseManager: base.NewBaseManager(configFile),
		config:      DefaultConfig(),
	}
}

// Ensure Manager implements base.ConfigManager interface
var _ base.ConfigManager = (*Manager)(nil)

// SetContext sets the Wails runtime context for emitting events
func (m *Manager) SetContext(ctx context.Context, log logger.Logger) {
	m.BaseManager.SetContext(ctx, log)
}

// Load loads the configuration from file
func (m *Manager) Load() error {
	mu := m.GetMutex()
	mu.Lock()
	defer mu.Unlock()

	// Ensure directory exists
	if err := base.EnsureDir(configFile); err != nil {
		return err
	}

	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		m.config = DefaultConfig()
		if err := m.saveLocked(); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		return nil
	}

	// Read config file
	data, err := os.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse config
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure version is set
	if config.Version == 0 {
		config.Version = CurrentVersion
	}

	m.config = &config
	return nil
}

// Get returns a copy of the current configuration
func (m *Manager) Get() interface{} {
	mu := m.GetMutex()
	mu.RLock()
	defer mu.RUnlock()

	// Return a copy to prevent external modifications
	configCopy := *m.config
	return &configCopy
}

// GetConfig returns the user config (type-safe version)
func (m *Manager) GetConfig() *Config {
	result := m.Get()
	if config, ok := result.(*Config); ok {
		return config
	}
	return DefaultConfig()
}

// Patch applies a partial update to the configuration
func (m *Manager) Patch(patch map[string]interface{}) error {
	mu := m.GetMutex()
	mu.Lock()
	defer mu.Unlock()

	if m.config == nil {
		return fmt.Errorf("config is not loaded")
	}

	// Use base helper for patching
	configMap, err := base.PatchConfig(m.config, patch)
	if err != nil {
		return err
	}

	// Convert back to Config struct
	var mergedConfig Config
	if err := base.UnmarshalPatchedConfig(configMap.(map[string]interface{}), &mergedConfig); err != nil {
		return err
	}

	// Ensure version is preserved
	mergedConfig.Version = m.config.Version

	// Update in-memory config
	m.config = &mergedConfig

	// Emit config:updated event for optimistic UI update
	m.EmitUpdatedWithName("config:updated", m.config)

	// Schedule save with debounce
	m.ScheduleSave(func() error {
		return m.save()
	}, "config")

	return nil
}

// Save saves the configuration to file
func (m *Manager) Save() error {
	mu := m.GetMutex()
	mu.Lock()
	defer mu.Unlock()

	if m.config == nil {
		return fmt.Errorf("config is not loaded")
	}

	return m.saveLocked()
}

// saveLocked saves the configuration to file (must be called with lock held)
func (m *Manager) saveLocked() error {
	return base.SaveJSONConfig(
		m.config,
		configFile,
		0o600,
		func() {
			if m.config.Version == 0 {
				m.config.Version = CurrentVersion
			}
		},
	)
}

// save saves the configuration to file (internal, assumes lock is held)
func (m *Manager) save() error {
	if m.config == nil {
		return fmt.Errorf("config is not loaded")
	}

	return m.saveLocked()
}
