package core

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"paperbox/internal/config/storage"

	"github.com/wailsapp/wails/v2/pkg/logger"
)

// BaseManager provides common functionality for configuration managers.
// It handles synchronization, debouncing, events, and storage operations.
type BaseManager[T any] struct {
	mu         sync.RWMutex
	debounce   *Debouncer
	events     *EventBus
	storage    storage.Storage
	config     *T
	configFile string
	eventName  string
	loader     func() (*T, error)
	validator  func(*T) error
	ensureFunc func(*T) // Function to ensure version and defaults
}

// BaseManagerOptions contains options for creating a BaseManager.
type BaseManagerOptions[T any] struct {
	Storage    storage.Storage
	ConfigFile string
	EventName  string
	Loader     func() (*T, error)
	Validator  func(*T) error
	EnsureFunc func(*T)
}

// NewBaseManager creates a new BaseManager with the provided options.
func NewBaseManager[T any](opts BaseManagerOptions[T]) *BaseManager[T] {
	return &BaseManager[T]{
		debounce:   NewDebouncer(DefaultDebounceDuration),
		events:     NewEventBus(context.TODO(), nil),
		storage:    opts.Storage,
		configFile: opts.ConfigFile,
		eventName:  opts.EventName,
		loader:     opts.Loader,
		validator:  opts.Validator,
		ensureFunc: opts.EnsureFunc,
	}
}

// SetContext sets the Wails runtime context for emitting events.
func (b *BaseManager[T]) SetContext(ctx context.Context, log logger.Logger) {
	b.events.SetContext(ctx, log)
}

// Load loads the configuration from storage.
func (b *BaseManager[T]) Load() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.loader != nil {
		// Use custom loader if provided
		cfg, err := b.loader()
		if err != nil {
			return err
		}
		b.config = cfg
		return nil
	}

	// Default loader: use storage
	var cfg T
	if err := b.storage.Load(b.configFile, &cfg); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Ensure defaults/version
	if b.ensureFunc != nil {
		b.ensureFunc(&cfg)
	}

	// Validate if validator is provided
	if b.validator != nil {
		if err := b.validator(&cfg); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}

	b.config = &cfg
	return nil
}

// Get returns a copy of the current configuration.
func (b *BaseManager[T]) Get() *T {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if b.config == nil {
		var zero T
		return &zero
	}

	// Return a deep copy to prevent external modifications
	return b.deepCopy(b.config)
}

// deepCopy creates a deep copy of the config using JSON marshaling.
func (b *BaseManager[T]) deepCopy(src *T) *T {
	data, err := json.Marshal(src)
	if err != nil {
		// If marshaling fails, return a shallow copy as fallback
		var dst T
		dst = *src
		return &dst
	}

	var dst T
	if err := json.Unmarshal(data, &dst); err != nil {
		// If unmarshaling fails, return a shallow copy as fallback
		dst = *src
		return &dst
	}

	return &dst
}

// Patch applies a partial update to the configuration.
func (b *BaseManager[T]) Patch(patch map[string]interface{}) error {
	ctx := b.events.Context()

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.config == nil {
		return fmt.Errorf("config is not loaded")
	}

	// Merge patch into current config
	var merged T
	if err := storage.MergePatch(b.config, patch, &merged); err != nil {
		return fmt.Errorf("failed to merge patch: %w", err)
	}

	// Ensure defaults/version
	if b.ensureFunc != nil {
		b.ensureFunc(&merged)
	}

	// Validate if validator is provided
	if b.validator != nil {
		if err := b.validator(&merged); err != nil {
			return fmt.Errorf("merged config validation failed: %w", err)
		}
	}

	// Update in-memory config
	b.config = &merged

	// Emit updated event
	if b.eventName != "" {
		b.events.Updated(b.eventName+":updated", b.config)
	}

	// Schedule save with debounce
	b.debounce.Schedule(func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if err := b.saveLocked(); err != nil {
			if ctx != nil {
				b.events.Error(b.eventName+":error", err.Error())
			}
		} else {
			if ctx != nil {
				b.events.Saved(b.eventName+":saved", b.configFile)
			}
		}
	})

	return nil
}

// Save saves the configuration to storage immediately (bypasses debounce).
func (b *BaseManager[T]) Save() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.config == nil {
		return fmt.Errorf("config is not loaded")
	}

	return b.saveLocked()
}

// saveLocked saves the configuration to storage (must be called with lock held).
func (b *BaseManager[T]) saveLocked() error {
	// Ensure defaults/version before saving
	if b.ensureFunc != nil {
		b.ensureFunc(b.config)
	}

	return b.storage.Save(b.configFile, b.config)
}

// UpdateConfig updates the in-memory configuration and schedules a save.
// This is useful for operations that modify the config directly.
func (b *BaseManager[T]) UpdateConfig(updater func(*T) error) error {
	ctx := b.events.Context()

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.config == nil {
		return fmt.Errorf("config is not loaded")
	}

	// Update config
	if err := updater(b.config); err != nil {
		return err
	}

	// Ensure defaults/version
	if b.ensureFunc != nil {
		b.ensureFunc(b.config)
	}

	// Validate if validator is provided
	if b.validator != nil {
		if err := b.validator(b.config); err != nil {
			return fmt.Errorf("config validation failed: %w", err)
		}
	}

	// Emit updated event
	if b.eventName != "" {
		b.events.Updated(b.eventName+":updated", b.config)
	}

	// Schedule save with debounce
	b.debounce.Schedule(func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if err := b.saveLocked(); err != nil {
			if ctx != nil {
				b.events.Error(b.eventName+":error", err.Error())
			}
		} else {
			if ctx != nil {
				b.events.Saved(b.eventName+":saved", b.configFile)
			}
		}
	})

	return nil
}

// GetConfig returns the current configuration (internal use, not a copy).
// This should only be used within the manager when lock is already held.
func (b *BaseManager[T]) GetConfig() *T {
	return b.config
}

// Events returns the EventBus for emitting custom events.
func (b *BaseManager[T]) Events() *EventBus {
	return b.events
}

