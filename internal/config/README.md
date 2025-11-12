# Config Package

This package manages all application configurations using a unified architecture.

## Architecture

The package uses a base manager pattern with interfaces to provide common functionality:

- **`base/`** - Base functionality for all config managers:
  - `manager.go` - BaseManager with common functionality (events, debounce, mutex)
  - `patch.go` - Helper functions for patching configs (marshal/unmarshal merge)
  - `save.go` - Helper function for saving JSON configs atomically
  - `file.go` - File operations (EnsureDir)
  - `writefile_*.go` - Atomic file writing (platform-specific)

- **`requests/`** - Requests configuration (HTTP requests and folders)
- **`user/`** - User configuration (theme, fontSize, baseURL)

## Adding New Configs

To add a new config:

1. Create a new subdirectory (e.g., `newconfig/`)
2. Define your config struct
3. Create a Manager that embeds `*base.BaseManager` and implements `base.ConfigManager`
4. Use base helpers:
   - `base.PatchConfig()` for patching
   - `base.SaveJSONConfig()` for saving JSON configs
   - `base.EnsureDir()` for directory creation
5. Add the config to `Manager` in `manager.go`

## Common Patterns

All config managers should:

- Embed `*base.BaseManager`
- Implement `base.ConfigManager` interface
- Use `base.PatchConfig()` for patching
- Use `base.SaveJSONConfig()` for saving JSON configs
- Use `base.EnsureDir()` for directory creation
- Emit events via `BaseManager.EmitUpdatedWithName()`
- Schedule saves via `BaseManager.ScheduleSave()`

## Example

```go
type Manager struct {
    *base.BaseManager
    config *MyConfig
}

func NewManager() *Manager {
    return &Manager{
        BaseManager: base.NewBaseManager(configFile),
    }
}

func (m *Manager) Patch(patch map[string]interface{}) error {
    mu := m.GetMutex()
    mu.Lock()
    defer mu.Unlock()

    configMap, err := base.PatchConfig(m.config, patch)
    // ... unmarshal and update
    m.BaseManager.ScheduleSave(func() error { return m.save() }, "myconfig")
    return nil
}
```
