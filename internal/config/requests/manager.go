package requests

import (
	"context"
	"fmt"
	"sync"

	"paperbox/internal/configutil"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Manager manages the requests configuration with in-memory state and debounced saves
type Manager struct {
	mu         sync.RWMutex
	storage    configutil.Storage
	events     *configutil.Events
	debounce   *configutil.Debounce
	config     *RequestsConfig
	configFile string
}

// getMapKeys returns a slice of keys from a map
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// NewManager creates a new requests config manager
func NewManager() *Manager {
	return &Manager{
		storage:    configutil.NewFileStorage(),
		events:     configutil.NewEvents(context.TODO()),
		debounce:   configutil.NewDebounce(configutil.DefaultDebounceDuration),
		configFile: getRequestsFilePath(),
	}
}

// NewManagerWithStorage creates a new requests config manager with custom storage (for testing)
func NewManagerWithStorage(storage configutil.Storage) *Manager {
	return &Manager{
		storage:    storage,
		events:     configutil.NewEvents(context.TODO()),
		debounce:   configutil.NewDebounce(configutil.DefaultDebounceDuration),
		configFile: getRequestsFilePath(),
	}
}

// getRequestsFilePath returns the path to the requests config file
func getRequestsFilePath() string {
	return requestsFile
}

// SetContext sets the Wails runtime context for emitting events
func (m *Manager) SetContext(ctx context.Context, log logger.Logger) {
	m.events.SetContext(ctx)
}

// Load loads the configuration from file
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cfg, err := Load()
	if err != nil {
		return err
	}

	m.config = cfg
	return nil
}

// Get returns a copy of the current configuration
func (m *Manager) Get() interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return NewRequestsConfig()
	}

	// Return a copy to prevent external modifications
	configCopy := *m.config
	valuesCopy := make(map[string]Item, len(m.config.Values))
	for k, v := range m.config.Values {
		valuesCopy[k] = v
	}
	configCopy.Values = valuesCopy

	return &configCopy
}

// GetRequestsConfig returns the requests config (type-safe version)
func (m *Manager) GetRequestsConfig() *RequestsConfig {
	result := m.Get()
	if cfg, ok := result.(*RequestsConfig); ok {
		return cfg
	}
	return NewRequestsConfig()
}

// PatchValues applies a partial update to the requests configuration using typed values
func (m *Manager) PatchValues(values map[string]Item) error {
	// Get context BEFORE locking to avoid deadlock
	ctx := m.events.GetContext()

	m.mu.Lock()
	defer m.mu.Unlock()

	if ctx != nil {
		runtime.LogInfo(ctx, fmt.Sprintf("PatchValues called with %d items", len(values)))
	}

	if m.config == nil {
		if ctx != nil {
			runtime.LogError(ctx, "PatchValues: config is not loaded")
		}
		return fmt.Errorf("config is not loaded")
	}

	// Create a copy of current config
	mergedConfig := *m.config
	if mergedConfig.Values == nil {
		mergedConfig.Values = make(map[string]Item)
	}

	// Merge values into config
	for k, v := range values {
		mergedConfig.Values[k] = v
	}

	// Ensure version is preserved
	mergedConfig.Version = m.config.Version

	// Validate merged config
	if err := Validate(&mergedConfig); err != nil {
		if ctx != nil {
			runtime.LogError(ctx, fmt.Sprintf("Validation failed: %v", err))
		}
		return fmt.Errorf("merged config validation failed: %w", err)
	}

	// Update in-memory config
	m.config = &mergedConfig

	if ctx != nil {
		runtime.LogInfo(ctx, fmt.Sprintf("Config updated in memory, values count: %d", len(m.config.Values)))
	}

	// Convert config to map for proper serialization
	eventData := map[string]interface{}{
		"version": m.config.Version,
		"values":  m.config.Values,
	}
	// Emit requests:updated event for optimistic UI update
	if ctx != nil {
		runtime.LogInfo(ctx, fmt.Sprintf("About to emit requests:updated event with %d items", len(m.config.Values)))
		runtime.EventsEmit(ctx, "requests:updated", eventData)
		runtime.LogInfo(ctx, "Event requests:updated emitted")
	}

	// Schedule save with debounce
	m.debounce.Schedule(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if err := m.saveLocked(); err != nil {
			if ctx != nil {
				m.events.EmitError("requests:error", err.Error())
			}
		} else {
			if ctx != nil {
				m.events.EmitSaved("requests:saved", m.configFile)
			}
		}
	})

	return nil
}

// AddRequest adds a new request to a parent folder
func (m *Manager) AddRequest(parentId string, name string, method string, path string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return "", fmt.Errorf("config is not loaded")
	}

	// Generate UUID
	newId := uuid.New().String()

	// Create new request item
	newItem := Item{
		Type:   ItemTypeRequest,
		Name:   name,
		Method: method,
		Path:   path,
	}

	// Get parent folder
	parent, exists := m.config.Values[parentId]
	if !exists || parent.Type != ItemTypeFolder {
		return "", fmt.Errorf("parent folder not found")
	}

	// Add new item to config
	m.config.Values[newId] = newItem

	// Add to parent's children
	children := make([]string, len(parent.Children))
	copy(children, parent.Children)
	children = append(children, newId)
	m.config.Values[parentId] = Item{
		Type:     parent.Type,
		Name:     parent.Name,
		Children: children,
	}

	// Emit updated event
	eventData := map[string]interface{}{
		"version": m.config.Version,
		"values":  m.config.Values,
	}
	m.events.EmitUpdated("requests:updated", eventData)

	// Schedule save with debounce
	ctx := m.events.GetContext()
	m.debounce.Schedule(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if err := m.saveLocked(); err != nil {
			if ctx != nil {
				m.events.EmitError("requests:error", err.Error())
			}
		} else {
			if ctx != nil {
				m.events.EmitSaved("requests:saved", m.configFile)
			}
		}
	})

	return newId, nil
}

// AddFolder adds a new folder to a parent folder
func (m *Manager) AddFolder(parentId string, name string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config == nil {
		return "", fmt.Errorf("config is not loaded")
	}

	// Generate UUID
	newId := uuid.New().String()

	// Create new folder item
	newItem := Item{
		Type:     ItemTypeFolder,
		Name:     name,
		Children: []string{},
	}

	// Get parent folder
	parent, exists := m.config.Values[parentId]
	if !exists || parent.Type != ItemTypeFolder {
		return "", fmt.Errorf("parent folder not found")
	}

	// Add new item to config
	m.config.Values[newId] = newItem

	// Add to parent's children
	children := make([]string, len(parent.Children))
	copy(children, parent.Children)
	children = append(children, newId)
	m.config.Values[parentId] = Item{
		Type:     parent.Type,
		Name:     parent.Name,
		Children: children,
	}

	// Emit updated event
	eventData := map[string]interface{}{
		"version": m.config.Version,
		"values":  m.config.Values,
	}
	m.events.EmitUpdated("requests:updated", eventData)

	// Schedule save with debounce
	ctx := m.events.GetContext()
	m.debounce.Schedule(func() {
		m.mu.Lock()
		defer m.mu.Unlock()
		if err := m.saveLocked(); err != nil {
			if ctx != nil {
				m.events.EmitError("requests:error", err.Error())
			}
		} else {
			if ctx != nil {
				m.events.EmitSaved("requests:saved", m.configFile)
			}
		}
	})

	return newId, nil
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
		0o644,
		func() {
			if m.config.Version == 0 {
				m.config.Version = CurrentVersion
			}
		},
	)
}
