package base

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bep/debounce"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	// DebounceDuration is the default duration to wait before saving config
	DebounceDuration = 700 * time.Millisecond
)

// ConfigManager defines the interface for configuration managers
type ConfigManager interface {
	// Load loads the configuration from file
	Load() error
	// Get returns a copy of the current configuration
	Get() interface{}
	// Patch applies a partial update to the configuration
	Patch(patch map[string]interface{}) error
	// SetContext sets the Wails runtime context and logger for emitting events
	SetContext(ctx context.Context, log logger.Logger)
	// Save saves the configuration to file (for manual saves)
	Save() error
}

// BaseManager provides common functionality for configuration managers
type BaseManager struct {
	mu            sync.RWMutex
	ctx           context.Context
	saveDebouncer func(func())
	configFile    string
}

// NewBaseManager creates a new base manager
func NewBaseManager(configFile string) *BaseManager {
	return &BaseManager{
		configFile:    configFile,
		saveDebouncer: debounce.New(DebounceDuration),
	}
}

// SetContext sets the Wails runtime context for emitting events
func (m *BaseManager) SetContext(ctx context.Context, log logger.Logger) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ctx = ctx
	// Logger is available through runtime.Log* functions with context
}

// GetContext returns the current context
func (m *BaseManager) GetContext() context.Context {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ctx
}

// GetConfigFile returns the config file path
func (m *BaseManager) GetConfigFile() string {
	return m.configFile
}

// ScheduleSave schedules a save operation with debounce
// eventPrefix can be used to customize event names (e.g., "requests", "config")
func (m *BaseManager) ScheduleSave(saveFunc func() error, eventPrefix ...string) {
	prefix := "config"
	if len(eventPrefix) > 0 && eventPrefix[0] != "" {
		prefix = eventPrefix[0]
	}

	m.saveDebouncer(func() {
		if err := saveFunc(); err != nil {
			m.EmitErrorWithName(prefix+":error", err.Error())
		} else {
			m.EmitSavedWithName(prefix + ":saved")
		}
	})
}

// EmitUpdated emits config:updated event
// eventName can be customized (e.g., "requests:updated", "config:updated")
func (m *BaseManager) EmitUpdated(config interface{}) {
	m.EmitUpdatedWithName("config:updated", config)
}

// EmitUpdatedWithName emits a custom updated event
func (m *BaseManager) EmitUpdatedWithName(eventName string, config interface{}) {
	fmt.Printf("[DEBUG] EmitUpdatedWithName called for event: %s\n", eventName)
	ctx := m.GetContext()
	fmt.Printf("[DEBUG] EmitUpdatedWithName: ctx is nil: %v\n", ctx == nil)
	if ctx == nil {
		fmt.Printf("[DEBUG] EmitUpdatedWithName: ctx is nil, returning early\n")
		return
	}
	runtime.LogInfo(ctx, fmt.Sprintf("Emitting event: %s", eventName))
	runtime.EventsEmit(ctx, eventName, config)
	runtime.LogInfo(ctx, fmt.Sprintf("Event %s emitted", eventName))
	fmt.Printf("[DEBUG] EmitUpdatedWithName: event %s emitted successfully\n", eventName)
}

// EmitSaved emits config:saved event
// eventName can be customized (e.g., "requests:saved", "config:saved")
func (m *BaseManager) EmitSaved() {
	m.EmitSavedWithName("config:saved")
}

// EmitSavedWithName emits a custom saved event
func (m *BaseManager) EmitSavedWithName(eventName string) {
	ctx := m.GetContext()
	if ctx == nil {
		return
	}

	fileInfo, err := os.Stat(m.configFile)
	updatedAt := time.Now()
	if err == nil {
		updatedAt = fileInfo.ModTime()
	}

	savedData := map[string]interface{}{
		"updatedAt": updatedAt.Format(time.RFC3339),
		"path":      m.configFile,
	}

	runtime.EventsEmit(ctx, eventName, savedData)
}

// EmitError emits config:error event
// eventName can be customized (e.g., "requests:error", "config:error")
func (m *BaseManager) EmitError(message string) {
	m.EmitErrorWithName("config:error", message)
}

// EmitErrorWithName emits a custom error event
func (m *BaseManager) EmitErrorWithName(eventName string, message string) {
	ctx := m.GetContext()
	if ctx == nil {
		return
	}

	errorData := map[string]interface{}{
		"message": message,
	}

	runtime.EventsEmit(ctx, eventName, errorData)
}

// GetMutex returns the mutex for synchronization
func (m *BaseManager) GetMutex() *sync.RWMutex {
	return &m.mu
}
