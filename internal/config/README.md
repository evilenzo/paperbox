# Config Package

This package manages all application configurations using a simplified architecture based on composition.

## Architecture

The package uses composition instead of inheritance, with utilities in a separate `configutil` package:

- **`configutil/`** - Utility functions for all config managers:
  - `storage.go` - Storage interface (WriteFileAtomic, PatchConfig) for dependency injection in tests
  - `events.go` - Wails event emission functionality
  - `debounce.go` - Debounced save functionality
  - `file.go` - File operations (EnsureDir, WriteFileAtomic - cross-platform)
  - `patch.go` - Helper functions for patching configs (marshal/unmarshal merge)
  - `save.go` - Helper function for saving JSON configs atomically

- **`requests/`** - Requests configuration (HTTP requests and folders)
- **`user/`** - User configuration (theme, fontSize, baseURL)
- **`interface.go`** - Common interface for all config managers
- **`manager.go`** - Main manager that aggregates all config managers

## Key Changes from Previous Architecture

1. **No inheritance**: Managers use composition with `configutil` utilities instead of embedding `BaseManager`
2. **Cross-platform file operations**: Single `WriteFileAtomic` implementation works on all platforms (no platform-specific files)
3. **Simpler structure**: Each manager composes only what it needs (storage, events, debounce)
4. **Dependency injection**: Storage interface allows injecting mock implementations for testing
5. **Direct mutex usage**: Uses `sync.RWMutex` directly instead of unnecessary wrapper

## Adding New Configs

To add a new config:

1. Create a new subdirectory (e.g., `newconfig/`)
2. Define your config struct
3. Create a Manager that implements `config.ManagerInterface` and uses `configutil` utilities:
   ```go
   type Manager struct {
       mu         sync.RWMutex
       storage    configutil.Storage
       events     *configutil.Events
       debounce   *configutil.Debounce
       config     *MyConfig
       configFile string
   }
   ```
4. Use storage interface and configutil helpers:
   - `storage.PatchConfig()` for patching
   - `storage.WriteFileAtomic()` for atomic file writes
   - `configutil.SaveJSONConfig(storage, ...)` for saving JSON configs
   - `configutil.EnsureDir()` for directory creation
   - `configutil.UnmarshalPatchedConfig()` for unmarshaling patched configs
5. Add the config to `Manager` in `manager.go`

## Common Patterns

All config managers should:

- Use `sync.RWMutex` directly for synchronization (no wrapper needed)
- Use composition with `configutil.Storage`, `configutil.Events`, and `configutil.Debounce`
- Implement `config.ManagerInterface` interface
- Use `storage.PatchConfig()` for patching
- Use `storage.WriteFileAtomic()` for atomic file writes
- Use `configutil.SaveJSONConfig(storage, ...)` for saving JSON configs
- Use `configutil.EnsureDir()` for directory creation
- Use `configutil.UnmarshalPatchedConfig()` for unmarshaling patched configs
- Emit events via `events.EmitUpdated()`, `events.EmitSaved()`, `events.EmitError()`
- Schedule saves via `debounce.Schedule()`
- Provide `NewManagerWithStorage()` constructor for testing with mock storage

## Example

```go
type Manager struct {
    mu         sync.RWMutex
    storage    configutil.Storage
    events     *configutil.Events
    debounce   *configutil.Debounce
    config     *MyConfig
    configFile string
}

func NewManager() *Manager {
    return &Manager{
        storage:    configutil.NewFileStorage(),
        events:     configutil.NewEvents(context.TODO()),
        debounce:   configutil.NewDebounce(configutil.DefaultDebounceDuration),
        configFile: getConfigFilePath(),
    }
}

func NewManagerWithStorage(storage configutil.Storage) *Manager {
    return &Manager{
        storage:    storage,
        events:     configutil.NewEvents(context.TODO()),
        debounce:   configutil.NewDebounce(configutil.DefaultDebounceDuration),
        configFile: getConfigFilePath(),
    }
}

func (m *Manager) Patch(patch map[string]interface{}) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    configMap, err := m.storage.PatchConfig(m.config, patch)
    if err != nil {
        return err
    }

    // Unmarshal patched config
    var mergedConfig MyConfig
    if err := configutil.UnmarshalPatchedConfig(configMap.(map[string]interface{}), &mergedConfig); err != nil {
        return err
    }

    m.config = &mergedConfig

    ctx := m.events.GetContext()
    m.debounce.Schedule(func() {
        m.mu.Lock()
        defer m.mu.Unlock()
        if err := m.saveLocked(); err != nil {
            if ctx != nil {
                m.events.EmitError("myconfig:error", err.Error())
            }
        } else {
            if ctx != nil {
                m.events.EmitSaved("myconfig:saved", m.configFile)
            }
        }
    })
    return nil
}

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
```
