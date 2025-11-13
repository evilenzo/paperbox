# Config Package

This package manages every persistent configuration in Paperbox. Each config type owns its own manager, file and synchronization primitives so that saving a large blob never blocks other configs.

## Layout

- **`infra/`** – runtime helpers shared by managers (event bus + debouncer).
- **`storage/`** – persistence primitives (atomic writer, JSON helpers, patching, path utilities).
- **`requests/`** – hierarchical HTTP request tree config.
- **`user/`** – user preferences (theme, font size, base URL).
- **`interface.go`** – interface implemented by every config manager.
- **`manager.go`** – aggregate that wires multiple configs into the app.

## Design Goals

1. **Multiple independent configs**: each manager works on its own file and mutex, so saving one does not stall the others.
2. **Compositional helpers**: managers compose the infra + storage helpers they need instead of relying on a big util package.
3. **Test-friendly persistence**: `storage.Writer` can be swapped for mocks while `storage.MergePatch` keeps patch logic close to persistence.
4. **Runtime integration**: `infra.EventBus` centralises Wails event emission and logging, while `infra.Debouncer` prevents excessive disk IO.

## Adding a New Config

1. Create a subpackage (e.g. `internal/config/foo`).
2. Define your config struct and defaults.
3. Create a manager that implements `config.ManagerInterface`:

```go
type Manager struct {
    mu         sync.RWMutex
    writer     storage.Writer
    events     *infra.EventBus
    debounce   *infra.Debouncer
    config     *FooConfig
    configFile string
}

func NewManager() *Manager {
    return &Manager{
        writer:     storage.NewFileWriter(),
        events:     infra.NewEventBus(context.TODO(), nil),
        debounce:   infra.NewDebouncer(infra.DefaultDebounceDuration),
        config:     DefaultFooConfig(),
        configFile: fooFilePath(),
    }
}
```

4. Use the helpers close to where they belong:
   - `storage.MergePatch(current, patch, &updated)` to apply map-based patches.
   - `storage.SaveJSON(writer, cfg, file, perm, ensureVersion)` for atomic writes.
   - `storage.EnsureParentDir(file)` when you need the directory to exist for first-run loads.
   - `events.Updated/Saved/Error` for UI notifications.
   - `debounce.Schedule` to delay disk writes when patching often.
5. Register the manager inside `config.NewManager()` (or expose it via your own accessor).

## Common Pattern

```go
func (m *Manager) Patch(patch map[string]interface{}) error {
    m.mu.Lock()
    defer m.mu.Unlock()

    var merged FooConfig
    if err := storage.MergePatch(m.config, patch, &merged); err != nil {
        return err
    }
    m.config = &merged
    m.events.Updated("foo:updated", m.config)

    ctx := m.events.Context()
    m.debounce.Schedule(func() {
        m.mu.Lock()
        defer m.mu.Unlock()
        if err := storage.SaveJSON(m.writer, m.config, m.configFile, 0o644, ensureVersion); err != nil {
            if ctx != nil {
                m.events.Error("foo:error", err.Error())
            }
            return
        }
        if ctx != nil {
            m.events.Saved("foo:saved", m.configFile)
        }
    })
    return nil
}
```
