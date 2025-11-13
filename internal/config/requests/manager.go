package requests

import (
	"context"
	"fmt"

	"paperbox/internal/config/core"
	"paperbox/internal/config/storage"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Manager manages the requests configuration with in-memory state and debounced saves
type Manager struct {
	*core.BaseManager[RequestsConfig]
}

// NewManager creates a new requests config manager
func NewManager(storage storage.Storage) *Manager {
	return &Manager{
		BaseManager: core.NewBaseManager(core.BaseManagerOptions[RequestsConfig]{
			Storage:    storage,
			ConfigFile: getRequestsFilePath(),
			EventName:  "requests",
			Loader:     Load,
			Validator:  Validate,
			EnsureFunc: func(cfg *RequestsConfig) {
				if cfg.Version == 0 {
					cfg.Version = CurrentVersion
				}
			},
		}),
	}
}

// NewManagerWithWriter creates a new requests config manager with a custom writer (for testing)
func NewManagerWithWriter(writer storage.Writer) *Manager {
	fileStorage := storage.NewFileStorageWithWriter(writer)
	coordinator := storage.NewStorageCoordinator(fileStorage, nil, nil)

	return &Manager{
		BaseManager: core.NewBaseManager(core.BaseManagerOptions[RequestsConfig]{
			Storage:    coordinator,
			ConfigFile: getRequestsFilePath(),
			EventName:  "requests",
			Loader:     Load,
			Validator:  Validate,
			EnsureFunc: func(cfg *RequestsConfig) {
				if cfg.Version == 0 {
					cfg.Version = CurrentVersion
				}
			},
		}),
	}
}

// getRequestsFilePath returns the path to the requests config file
func getRequestsFilePath() string {
	return requestsFile
}

// SetContext sets the Wails runtime context for emitting events
func (m *Manager) SetContext(ctx context.Context, log logger.Logger) {
	m.BaseManager.SetContext(ctx, log)
}

// Get returns a copy of the current configuration (implements ManagerInterface)
func (m *Manager) Get() interface{} {
	return m.GetRequestsConfig()
}

// GetRequestsConfig returns the requests config (type-safe version)
func (m *Manager) GetRequestsConfig() *RequestsConfig {
	return m.BaseManager.Get()
}

// PatchValues applies a partial update to the requests configuration using typed values
func (m *Manager) PatchValues(values map[string]Item) error {
	ctx := m.Events().Context()

	if ctx != nil {
		runtime.LogInfo(ctx, fmt.Sprintf("PatchValues called with %d items", len(values)))
	}

	return m.UpdateConfig(func(cfg *RequestsConfig) error {
		if cfg.Values == nil {
			cfg.Values = make(map[string]Item)
		}

		// Merge values into config
		for k, v := range values {
			cfg.Values[k] = v
		}

		if ctx != nil {
			runtime.LogInfo(ctx, fmt.Sprintf("Config updated in memory, values count: %d", len(cfg.Values)))
		}

		// Emit event with proper format
		eventData := map[string]interface{}{
			"version": cfg.Version,
			"values":  cfg.Values,
		}
		if ctx != nil {
			runtime.LogInfo(ctx, fmt.Sprintf("About to emit requests:updated event with %d items", len(cfg.Values)))
		}
		m.Events().Updated("requests:updated", eventData)
		if ctx != nil {
			runtime.LogInfo(ctx, "Event requests:updated emitted")
		}

		return nil
	})
}

// AddRequest adds a new request to a parent folder
func (m *Manager) AddRequest(parentId string, name string, method string, path string) (string, error) {
	var newId string

	err := m.UpdateConfig(func(cfg *RequestsConfig) error {
		// Generate UUID
		newId = uuid.New().String()

		// Create new request item
		newItem := Item{
			Type:   ItemTypeRequest,
			Name:   name,
			Method: method,
			Path:   path,
		}

		// Get parent folder
		parent, exists := cfg.Values[parentId]
		if !exists || parent.Type != ItemTypeFolder {
			return fmt.Errorf("parent folder not found")
		}

		// Add new item to config
		if cfg.Values == nil {
			cfg.Values = make(map[string]Item)
		}
		cfg.Values[newId] = newItem

		// Initialize parent.Children if nil
		if parent.Children == nil {
			parent.Children = []string{}
		}
		// Add to parent's children
		parent.Children = append(parent.Children, newId)
		cfg.Values[parentId] = parent

		// Emit updated event
		eventData := map[string]interface{}{
			"version": cfg.Version,
			"values":  cfg.Values,
		}
		m.Events().Updated("requests:updated", eventData)

		return nil
	})

	return newId, err
}

// AddFolder adds a new folder to a parent folder
func (m *Manager) AddFolder(parentId string, name string) (string, error) {
	var newId string

	err := m.UpdateConfig(func(cfg *RequestsConfig) error {
		// Generate UUID
		newId = uuid.New().String()

		// Create new folder item
		newItem := Item{
			Type:     ItemTypeFolder,
			Name:     name,
			Children: []string{},
		}

		// Get parent folder
		parent, exists := cfg.Values[parentId]
		if !exists || parent.Type != ItemTypeFolder {
			return fmt.Errorf("parent folder not found")
		}

		// Add new item to config
		if cfg.Values == nil {
			cfg.Values = make(map[string]Item)
		}
		cfg.Values[newId] = newItem

		// Initialize parent.Children if nil
		if parent.Children == nil {
			parent.Children = []string{}
		}
		// Add to parent's children
		parent.Children = append(parent.Children, newId)
		cfg.Values[parentId] = parent

		// Emit updated event
		eventData := map[string]interface{}{
			"version": cfg.Version,
			"values":  cfg.Values,
		}
		m.Events().Updated("requests:updated", eventData)

		return nil
	})

	return newId, err
}

// AddRootFolder adds a new root-level folder (without parent)
func (m *Manager) AddRootFolder(name string) (string, error) {
	var newId string

	err := m.UpdateConfig(func(cfg *RequestsConfig) error {
		// Generate UUID
		newId = uuid.New().String()

		// Create new folder item
		newItem := Item{
			Type:     ItemTypeFolder,
			Name:     name,
			Children: []string{},
		}

		// Add new item to config
		if cfg.Values == nil {
			cfg.Values = make(map[string]Item)
		}
		cfg.Values[newId] = newItem

		// Emit updated event
		eventData := map[string]interface{}{
			"version": cfg.Version,
			"values":  cfg.Values,
		}
		m.Events().Updated("requests:updated", eventData)

		return nil
	})

	return newId, err
}

// DeleteItem deletes an item from the requests configuration
func (m *Manager) DeleteItem(itemId string) error {
	return m.UpdateConfig(func(cfg *RequestsConfig) error {
		// Get item to delete
		item, exists := cfg.Values[itemId]
		if !exists {
			return fmt.Errorf("item not found")
		}

		// Remove from parent's children
		for parentId, parent := range cfg.Values {
			if parent.Type == ItemTypeFolder && parent.Children != nil {
				// Filter out the deleted item from children
				newChildren := []string{}
				for _, childId := range parent.Children {
					if childId != itemId {
						newChildren = append(newChildren, childId)
					}
				}
				// Only update if children changed
				if len(newChildren) != len(parent.Children) {
					parent.Children = newChildren
					cfg.Values[parentId] = parent
				}
			}
		}

		// If it's a folder, also delete all children recursively
		if item.Type == ItemTypeFolder && item.Children != nil {
			for _, childId := range item.Children {
				// Recursively delete children (but don't call DeleteItem to avoid nested UpdateConfig)
				delete(cfg.Values, childId)
			}
		}

		// Delete the item itself
		delete(cfg.Values, itemId)

		// Emit updated event
		eventData := map[string]interface{}{
			"version": cfg.Version,
			"values":  cfg.Values,
		}
		m.Events().Updated("requests:updated", eventData)

		return nil
	})
}
