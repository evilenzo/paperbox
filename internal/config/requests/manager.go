package requests

import (
	"context"
	"fmt"

	"paperbox/internal/config/base"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Manager manages the requests configuration with in-memory state and debounced saves
type Manager struct {
	*base.BaseManager
	config *RequestsConfig
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
		BaseManager: base.NewBaseManager(getRequestsFilePath()),
	}
}

// getRequestsFilePath returns the path to the requests config file
func getRequestsFilePath() string {
	return requestsFile
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

	config, err := Load()
	if err != nil {
		return err
	}

	m.config = config
	return nil
}

// Get returns a copy of the current configuration
func (m *Manager) Get() interface{} {
	mu := m.GetMutex()
	mu.RLock()
	defer mu.RUnlock()

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
	if config, ok := result.(*RequestsConfig); ok {
		return config
	}
	return NewRequestsConfig()
}

// Patch applies a partial update to the configuration
func (m *Manager) Patch(patch map[string]interface{}) error {
	// Log immediately to verify method is called
	fmt.Printf("[DEBUG] Patch method called in requests manager\n")

	// Get context BEFORE locking to avoid deadlock (GetContext uses RLock on the same mutex)
	ctx := m.GetContext()
	fmt.Printf("[DEBUG] Patch: ctx is nil: %v\n", ctx == nil)

	mu := m.GetMutex()
	fmt.Printf("[DEBUG] Patch: about to lock mutex\n")
	mu.Lock()
	fmt.Printf("[DEBUG] Patch: mutex locked\n")
	defer mu.Unlock()
	if ctx != nil {
		fmt.Printf("[DEBUG] Patch: ctx is not nil, logging via runtime\n")
		runtime.LogInfo(ctx, fmt.Sprintf("Patch called with patch keys: %v", getMapKeys(patch)))
		// Log patch values structure
		if values, ok := patch["values"].(map[string]interface{}); ok {
			fmt.Printf("[DEBUG] Patch: values is map[string]interface{}, contains %d items\n", len(values))
			runtime.LogInfo(ctx, fmt.Sprintf("Patch values contains %d items", len(values)))
			// Log first few keys
			count := 0
			for key := range values {
				if count < 3 {
					runtime.LogInfo(ctx, fmt.Sprintf("Patch value key: %s", key))
					count++
				}
			}
		} else {
			fmt.Printf("[DEBUG] Patch: values is NOT map[string]interface{}\n")
			runtime.LogError(ctx, "Patch values is not a map[string]interface{}")
		}
	} else {
		fmt.Printf("[DEBUG] Patch: ctx is nil, skipping runtime logging\n")
	}

	if m.config == nil {
		if ctx != nil {
			runtime.LogError(ctx, "Patch: config is not loaded")
		}
		fmt.Printf("[DEBUG] Patch: config is nil\n")
		return fmt.Errorf("config is not loaded")
	}
	fmt.Printf("[DEBUG] Patch: config loaded, proceeding with patch\n")

	// Use base helper for patching
	configMap, err := base.PatchConfig(m.config, patch)
	if err != nil {
		if ctx != nil {
			runtime.LogError(ctx, fmt.Sprintf("Patch failed: %v", err))
		}
		return err
	}

	// Convert back to RequestsConfig struct
	var mergedConfig RequestsConfig
	if err := base.UnmarshalPatchedConfig(configMap.(map[string]interface{}), &mergedConfig); err != nil {
		if ctx != nil {
			runtime.LogError(ctx, fmt.Sprintf("UnmarshalPatchedConfig failed: %v", err))
		}
		return err
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
	// Use ctx directly to avoid deadlock (EmitUpdatedWithName would call GetContext() which needs RLock)
	if ctx != nil {
		runtime.LogInfo(ctx, fmt.Sprintf("About to emit requests:updated event with %d items", len(m.config.Values)))
		// Log a sample item to verify it's updated
		for id, item := range m.config.Values {
			runtime.LogInfo(ctx, fmt.Sprintf("Sample item in event: %s, name=\"%s\"", id, item.Name))
			break // Only log first item
		}
		runtime.LogInfo(ctx, "Emitting event: requests:updated")
		runtime.EventsEmit(ctx, "requests:updated", eventData)
		runtime.LogInfo(ctx, "Event requests:updated emitted")
	} else {
		fmt.Printf("[DEBUG] Patch: ctx is nil, cannot emit event\n")
	}

	// Schedule save with debounce
	m.ScheduleSave(func() error {
		return m.save()
	}, "requests")

	return nil
}

// AddRequest adds a new request to a parent folder
func (m *Manager) AddRequest(parentId string, name string, method string, path string) (string, error) {
	mu := m.GetMutex()
	mu.Lock()
	defer mu.Unlock()

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
	m.EmitUpdatedWithName("requests:updated", m.config)

	// Schedule save with debounce
	m.ScheduleSave(func() error {
		return m.save()
	}, "requests")

	return newId, nil
}

// AddFolder adds a new folder to a parent folder
func (m *Manager) AddFolder(parentId string, name string) (string, error) {
	mu := m.GetMutex()
	mu.Lock()
	defer mu.Unlock()

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
	m.EmitUpdatedWithName("requests:updated", m.config)

	// Schedule save with debounce
	m.ScheduleSave(func() error {
		return m.save()
	}, "requests")

	return newId, nil
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
		getRequestsFilePath(),
		0o644,
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
