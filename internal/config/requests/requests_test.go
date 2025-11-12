package requests

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *RequestsConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config with requests and folders",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"req1": {
						Type:   ItemTypeRequest,
						Name:   "Get Users",
						Method: "GET",
						Path:   "/api/users",
					},
					"folder1": {
						Type:     ItemTypeFolder,
						Name:     "API",
						Children: []string{"req1"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "request with children should fail",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"req1": {
						Type:     ItemTypeRequest,
						Name:     "Get Users",
						Method:   "GET",
						Path:     "/api/users",
						Children: []string{"req2"},
					},
				},
			},
			wantErr: true,
			errMsg:  "request cannot have children",
		},
		{
			name: "folder with method should fail",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"folder1": {
						Type:   ItemTypeFolder,
						Name:   "API",
						Method: "GET",
					},
				},
			},
			wantErr: true,
			errMsg:  "folder cannot have a method",
		},
		{
			name: "folder with path should fail",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"folder1": {
						Type: ItemTypeFolder,
						Name: "API",
						Path: "/api",
					},
				},
			},
			wantErr: true,
			errMsg:  "folder cannot have a path",
		},
		{
			name: "request without method should fail",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"req1": {
						Type: ItemTypeRequest,
						Name: "Get Users",
						Path: "/api/users",
					},
				},
			},
			wantErr: true,
			errMsg:  "request must have a method",
		},
		{
			name: "request without path should fail",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"req1": {
						Type:   ItemTypeRequest,
						Name:   "Get Users",
						Method: "GET",
					},
				},
			},
			wantErr: true,
			errMsg:  "request must have a path",
		},
		{
			name: "invalid HTTP method should fail",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"req1": {
						Type:   ItemTypeRequest,
						Name:   "Get Users",
						Method: "INVALID",
						Path:   "/api/users",
					},
				},
			},
			wantErr: true,
			errMsg:  "http method",
		},
		{
			name: "missing child reference should fail",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"folder1": {
						Type:     ItemTypeFolder,
						Name:     "API",
						Children: []string{"nonexistent"},
					},
				},
			},
			wantErr: true,
			errMsg:  "child reference 'nonexistent' does not exist",
		},
		{
			name: "circular reference should fail",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"folder1": {
						Type:     ItemTypeFolder,
						Name:     "API",
						Children: []string{"folder1"},
					},
				},
			},
			wantErr: true,
			errMsg:  "circular reference detected",
		},
		{
			name: "empty name should fail",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"req1": {
						Type:   ItemTypeRequest,
						Name:   "",
						Method: "GET",
						Path:   "/api/users",
					},
				},
			},
			wantErr: true,
			errMsg:  "name is required",
		},
		{
			name: "invalid item type should fail",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"item1": {
						Type: ItemType("invalid"),
						Name: "Test",
					},
				},
			},
			wantErr: true,
			errMsg:  "type must be one of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errMsg != "" {
				if err == nil || err.Error() == "" {
					t.Errorf("Validate() expected error message containing '%s', got nil", tt.errMsg)
					return
				}
				// Check if error message contains the expected text (case-insensitive)
				errLower := strings.ToLower(err.Error())
				msgLower := strings.ToLower(tt.errMsg)
				if !contains(errLower, msgLower) {
					t.Errorf("Validate() error message = %v, want containing '%s'", err, tt.errMsg)
				}
			}
		})
	}
}

func TestMigrateConfig(t *testing.T) {
	tests := []struct {
		name            string
		config          *RequestsConfig
		expectedVersion int
	}{
		{
			name: "version 0 should migrate to 1",
			config: &RequestsConfig{
				Version: 0,
				Values: map[string]Item{
					"req1": {
						Type:   ItemTypeRequest,
						Name:   "Test",
						Method: "GET",
						Path:   "/test",
					},
				},
			},
			expectedVersion: 1,
		},
		{
			name: "version 1 should stay at 1",
			config: &RequestsConfig{
				Version: 1,
				Values: map[string]Item{
					"req1": {
						Type:   ItemTypeRequest,
						Name:   "Test",
						Method: "GET",
						Path:   "/test",
					},
				},
			},
			expectedVersion: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := migrateConfig(tt.config)
			if err != nil {
				t.Errorf("migrateConfig() error = %v", err)
				return
			}
			if tt.config.Version != tt.expectedVersion {
				t.Errorf("migrateConfig() version = %v, want %v", tt.config.Version, tt.expectedVersion)
			}
		})
	}
}

func TestLoadAndSaveConfig(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	originalAppDataDir := appDataDir
	appDataDir = tmpDir
	requestsFile = filepath.Join(tmpDir, RequestsFileName)
	defer func() {
		appDataDir = originalAppDataDir
		requestsFile = filepath.Join(appDataDir, RequestsFileName)
	}()

	// Test creating new config
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if config == nil {
		t.Fatal("Load() returned nil config")
	}
	if config.Version != CurrentVersion {
		t.Errorf("Load() version = %v, want %v", config.Version, CurrentVersion)
	}

	// Test saving and loading config
	testConfig := &RequestsConfig{
		Version: CurrentVersion,
		Values: map[string]Item{
			"req1": {
				Type:   ItemTypeRequest,
				Name:   "Test Request",
				Method: "GET",
				Path:   "/test",
			},
			"folder1": {
				Type:     ItemTypeFolder,
				Name:     "Test Folder",
				Children: []string{"req1"},
			},
		},
	}

	if err := Validate(testConfig); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if err := Save(testConfig); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loadedConfig, err := Load()
	if err != nil {
		t.Fatalf("Load() after save error = %v", err)
	}

	if loadedConfig.Version != testConfig.Version {
		t.Errorf("Load() version = %v, want %v", loadedConfig.Version, testConfig.Version)
	}

	if len(loadedConfig.Values) != len(testConfig.Values) {
		t.Errorf("Load() values count = %v, want %v", len(loadedConfig.Values), len(testConfig.Values))
	}

	// Verify request
	req1, exists := loadedConfig.Values["req1"]
	if !exists {
		t.Error("Load() req1 not found")
	} else {
		if req1.Type != ItemTypeRequest || req1.Name != "Test Request" || req1.Method != "GET" || req1.Path != "/test" {
			t.Errorf("Load() req1 = %+v, want different values", req1)
		}
	}

	// Verify folder
	folder1, exists := loadedConfig.Values["folder1"]
	if !exists {
		t.Error("Load() folder1 not found")
	} else {
		if folder1.Type != ItemTypeFolder || folder1.Name != "Test Folder" || len(folder1.Children) != 1 || folder1.Children[0] != "req1" {
			t.Errorf("Load() folder1 = %+v, want different values", folder1)
		}
	}
}

func TestConfigVersioning(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()
	originalAppDataDir := appDataDir
	appDataDir = tmpDir
	requestsFile = filepath.Join(tmpDir, RequestsFileName)
	defer func() {
		appDataDir = originalAppDataDir
		requestsFile = filepath.Join(appDataDir, RequestsFileName)
	}()

	// Test loading config without version (old format)
	oldFormatJSON := `{
		"values": {
			"req1": {
				"type": "request",
				"name": "Test",
				"method": "GET",
				"path": "/test"
			}
		}
	}`

	if err := os.WriteFile(requestsFile, []byte(oldFormatJSON), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if config.Version != CurrentVersion {
		t.Errorf("Load() migrated version = %v, want %v", config.Version, CurrentVersion)
	}

	// Verify data is preserved
	if len(config.Values) != 1 {
		t.Errorf("Load() values count = %v, want 1", len(config.Values))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestJSONUnmarshal(t *testing.T) {
	// Test that JSON unmarshaling works correctly
	jsonData := `{
		"version": 1,
		"values": {
			"req1": {
				"type": "request",
				"name": "Test Request",
				"method": "GET",
				"path": "/test"
			},
			"folder1": {
				"type": "folder",
				"name": "Test Folder",
				"children": ["req1"]
			}
		}
	}`

	var config RequestsConfig
	if err := json.Unmarshal([]byte(jsonData), &config); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if config.Version != 1 {
		t.Errorf("json.Unmarshal() version = %v, want 1", config.Version)
	}

	if len(config.Values) != 2 {
		t.Errorf("json.Unmarshal() values count = %v, want 2", len(config.Values))
	}
}
