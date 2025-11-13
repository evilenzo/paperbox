package configutil

import (
	"context"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Events provides Wails event emission functionality
type Events struct {
	ctx context.Context
}

// GetContext returns the current context (for internal use)
func (e *Events) GetContext() context.Context {
	return e.ctx
}

// NewEvents creates a new Events instance
func NewEvents(ctx context.Context) *Events {
	return &Events{ctx: ctx}
}

// SetContext sets the Wails runtime context
func (e *Events) SetContext(ctx context.Context) {
	e.ctx = ctx
}

// EmitUpdated emits an updated event
func (e *Events) EmitUpdated(eventName string, config interface{}) {
	if e.ctx == nil {
		return
	}
	runtime.EventsEmit(e.ctx, eventName, config)
}

// EmitSaved emits a saved event with file metadata
func (e *Events) EmitSaved(eventName string, configFile string) {
	if e.ctx == nil {
		return
	}

	fileInfo, err := os.Stat(configFile)
	updatedAt := time.Now()
	if err == nil {
		updatedAt = fileInfo.ModTime()
	}

	savedData := map[string]interface{}{
		"updatedAt": updatedAt.Format(time.RFC3339),
		"path":      configFile,
	}

	runtime.EventsEmit(e.ctx, eventName, savedData)
}

// EmitError emits an error event
func (e *Events) EmitError(eventName string, message string) {
	if e.ctx == nil {
		return
	}

	errorData := map[string]interface{}{
		"message": message,
	}

	runtime.EventsEmit(e.ctx, eventName, errorData)
}

