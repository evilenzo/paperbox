package config

import (
	"context"

	"github.com/wailsapp/wails/v2/pkg/logger"
)

// ManagerInterface defines the interface for configuration managers
type ManagerInterface interface {
	// Load loads the configuration from file
	Load() error
	// Get returns a copy of the current configuration
	Get() interface{}
	// SetContext sets the Wails runtime context and logger for emitting events
	SetContext(ctx context.Context, log logger.Logger)
	// Save saves the configuration to file (for manual saves)
	Save() error
}
