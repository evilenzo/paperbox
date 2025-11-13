package core

import (
	"context"
	"os"
	"time"

	"github.com/wailsapp/wails/v2/pkg/logger"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// EventBus knows how to talk to the Wails runtime and emit config related events.
type EventBus struct {
	ctx context.Context
	log logger.Logger
}

// NewEventBus builds a bus with optional runtime context and logger.
func NewEventBus(ctx context.Context, log logger.Logger) *EventBus {
	return &EventBus{ctx: ctx, log: log}
}

// SetContext wires the bus to the Wails runtime.
func (b *EventBus) SetContext(ctx context.Context, log logger.Logger) {
	b.ctx = ctx
	b.log = log
}

// Context returns the runtime context (used when code needs to log directly).
func (b *EventBus) Context() context.Context {
	return b.ctx
}

// Updated notifies the UI that config data changed in memory.
func (b *EventBus) Updated(event string, payload interface{}) {
	if b.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(b.ctx, event, payload)
}

// Saved notifies the UI that a config was flushed to disk (with metadata).
func (b *EventBus) Saved(event string, filePath string) {
	if b.ctx == nil {
		return
	}

	fileInfo, err := os.Stat(filePath)
	updatedAt := time.Now()
	if err == nil {
		updatedAt = fileInfo.ModTime()
	}

	wailsruntime.EventsEmit(b.ctx, event, map[string]interface{}{
		"updatedAt": updatedAt.Format(time.RFC3339),
		"path":      filePath,
	})
}

// Error notifies listeners about a persistence failure.
func (b *EventBus) Error(event string, message string) {
	if b.ctx == nil {
		return
	}

	wailsruntime.EventsEmit(b.ctx, event, map[string]interface{}{
		"message": message,
	})
}

