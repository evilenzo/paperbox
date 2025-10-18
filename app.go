package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path"

	"github.com/adrg/xdg"
)

var appDataDir = path.Join(xdg.DataHome, "rivulus")
var configFile = path.Join(appDataDir, "config.json")

type Request struct {
	Name string `json:"name"`
}

type RequestNode struct {
	Name     string        `json:"name"`
	Method   string        `json:"method"`
	Children []RequestNode `json:"children,omitempty"`
}

type Config struct {
	Requests []RequestNode `json:"requests"`
}

// App struct
type App struct {
	ctx    context.Context
	config Config
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// Create app data directory if it doesn't exist.
	if _, err := os.Stat(appDataDir); os.IsNotExist(err) {
		err := os.MkdirAll(appDataDir, 0755)
		if err != nil {
			log.Fatalf("Failed to create app data directory: %v", err)
		}
	}

	// Create config file if it doesn't exist.
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		err := os.WriteFile(configFile, []byte("{}"), 0644)
		if err != nil {
			log.Fatalf("Failed to create config file: %v", err)
		}
	}

	// Read config file.
	configFileContent, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	var config Config
	err = json.Unmarshal(configFileContent, &config)

	a.config = config

	if err != nil {
		log.Fatalf("Failed to unmarshal config file: %v", err)
	}
}

// Greet returns a greeting for the given name
func (a *App) Greet() []RequestNode {
	return a.config.Requests
}
