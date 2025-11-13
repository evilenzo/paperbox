package configutil

import (
	"time"

	"github.com/bep/debounce"
)

const (
	// DefaultDebounceDuration is the default duration to wait before saving config
	DefaultDebounceDuration = 700 * time.Millisecond
)

// Debounce provides debounced save functionality
type Debounce struct {
	debouncer func(func())
}

// NewDebounce creates a new debounce instance
func NewDebounce(duration time.Duration) *Debounce {
	return &Debounce{
		debouncer: debounce.New(duration),
	}
}

// Schedule schedules a function to be called after debounce delay
func (d *Debounce) Schedule(fn func()) {
	d.debouncer(fn)
}

