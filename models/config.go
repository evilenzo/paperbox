package models

// Config represents the user configuration for Wails bindings
type Config struct {
	Version  int    `json:"version"`
	Theme    string `json:"theme"`    // "light" | "dark" | "auto"
	FontSize int    `json:"fontSize"` // Font size in pixels
	BaseURL  string `json:"baseURL"`  // Base URL for API requests
}

