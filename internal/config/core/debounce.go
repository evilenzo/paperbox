package core

import (
	"time"

	"github.com/bep/debounce"
)

const (
	// DefaultDebounceDuration prevents disk writes on every keystroke.
	DefaultDebounceDuration = 700 * time.Millisecond
)

// Debouncer wraps the debounce helper so managers do not depend on globals.
type Debouncer struct {
	debounced func(func())
}

// NewDebouncer returns a callable debouncer with the provided delay.
func NewDebouncer(duration time.Duration) *Debouncer {
	return &Debouncer{debounced: debounce.New(duration)}
}

// Schedule triggers the callback after the debounce window.
func (d *Debouncer) Schedule(fn func()) {
	d.debounced(fn)
}
