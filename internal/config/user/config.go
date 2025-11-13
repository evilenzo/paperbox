package user

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"

	"paperbox/internal/configutil"

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
	mu         sync.RWMutex
	storage    configutil.Storage
	events     *configutil.Events
	debounce   *configutil.Debounce
	config     *Config
	configFile string
}

// NewManager creates a new config manager
func NewManager() *Manager {
	return &Manager{
		storage:    configutil.NewFileStorage(),
		events:     configutil.NewEvents(context.TODO()),
		debounce:   configutil.NewDebounce(configutil.DefaultDebounceDuration),
		config:     DefaultConfig(),
		configFile: configFile,
	}
}

// NewManagerWithStorage creates a new config manager with custom storage (for testing)
func NewManagerWithStorage(storage configutil.Storage) *Manager {
	return &Manager{
		storage:    storage,
		events:     configutil.NewEvents(context.TODO()),
		debounce:   configutil.NewDebounce(configutil.DefaultDebounceDuration),
		config:     DefaultConfig(),
		configFile: configFile,
	}
}

// SetContext sets the Wails runtime context for emitting events
func (m *Manager) SetContext(ctx context.Context, log logger.Logger) {
	m.events.SetContext(ctx)
}

// Load loads the configuration from file
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure directory exists
	if err := configutil.EnsureDir(configFile); err != nil {
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
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure version is set
	if cfg.Version == 0 {
		cfg.Version = CurrentVersion
	}

	m.config = &cfg
	return nil
}

// Get returns a copy of the current configuration
func (m *Manager) Get() interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to prevent external modifications
	configCopy := *m.config
	return &configCopy
}

// GetConfig returns the user config (type-safe version)
func (m *Manager) GetConfig() *Config {
	result := m.Get()
	if cfg, ok := result.(*Config); ok {
		return cfg
	}
	return DefaultConfig()
}

// Patch applies a partial update to the configuration
func (m *Manager) Patch(patch map[string]interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return fmt.Errorf("config is not loaded")
	}

	// Use storage helper for patching
	configMap, err := m.storage.PatchConfig(m.config, patch)
	if err != nil {
		return err
	}

	// Convert back to Config struct
	var mergedConfig Config
	if err := configutil.UnmarshalPatchedConfig(configMap.(map[string]interface{}), &mergedConfig); err != nil {
		return err
	}

	// Ensure version is preserved
	mergedConfig.Version = m.config.Version

	// Update in-memory config
	m.config = &mergedConfig

	// Emit config:updated event for optimistic UI update
	m.events.EmitUpdated("config:updated", m.config)

	// Schedule save with debounce
	ctx := m.events.GetContext()
	m.debounce.Schedule(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if err := m.saveLocked(); err != nil {
			if ctx != nil {
				m.events.EmitError("config:error", err.Error())
			}
		} else {
			if ctx != nil {
				m.events.EmitSaved("config:saved", m.configFile)
			}
		}
	})

	return nil
}

// Save saves the configuration to file
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return fmt.Errorf("config is not loaded")
	}

	return m.saveLocked()
}

// saveLocked saves the configuration to file (must be called with lock held)
func (m *Manager) saveLocked() error {
	return configutil.SaveJSONConfig(
		m.storage,
		m.config,
		m.configFile,
		0o600,
		func() {
			if m.config.Version == 0 {
				m.config.Version = CurrentVersion
			}
		},
	)
}
