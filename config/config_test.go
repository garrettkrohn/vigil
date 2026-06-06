package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		wantErr     bool
		wantServices int
	}{
		{
			name: "should load valid YAML config with services",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				content := `services:
  - name: "api-service"
    ports: [8080]
  - name: "auth-service"
    ports: [8081, 9091]
`
				if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return configPath
			},
			wantErr:      false,
			wantServices: 2,
		},
		{
			name: "should handle missing config file (create default)",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				return filepath.Join(tmpDir, "nonexistent.yaml")
			},
			wantErr:      false,
			wantServices: 2, // Default config has 2 example services
		},
		{
			name: "should return error for invalid YAML",
			setup: func(t *testing.T) string {
				tmpDir := t.TempDir()
				configPath := filepath.Join(tmpDir, "config.yaml")
				content := `services:
  - name: "broken
    ports: [8080
`
				if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
				return configPath
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := tt.setup(t)
			config, err := LoadConfigFromPath(configPath)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && config != nil && len(config.Services) != tt.wantServices {
				t.Errorf("expected %d services, got %d", tt.wantServices, len(config.Services))
			}
		})
	}
}

func TestParseService(t *testing.T) {
	tests := []struct {
		name      string
		yaml      string
		wantName  string
		wantPorts int
	}{
		{
			name: "should parse service with single port",
			yaml: `services:
  - name: "api"
    ports: [8080]
`,
			wantName:  "api",
			wantPorts: 1,
		},
		{
			name: "should parse service with multiple ports",
			yaml: `services:
  - name: "gateway"
    ports: [3000, 3001, 3002]
`,
			wantName:  "gateway",
			wantPorts: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			configPath := filepath.Join(tmpDir, "config.yaml")
			if err := os.WriteFile(configPath, []byte(tt.yaml), 0644); err != nil {
				t.Fatal(err)
			}

			config, err := LoadConfigFromPath(configPath)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(config.Services) == 0 {
				t.Fatal("expected at least one service")
			}

			service := config.Services[0]
			if service.Name != tt.wantName {
				t.Errorf("expected name %s, got %s", tt.wantName, service.Name)
			}
			if len(service.Ports) != tt.wantPorts {
				t.Errorf("expected %d ports, got %d", tt.wantPorts, len(service.Ports))
			}
		})
	}
}

func TestExpandHomeDir(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		wantHome bool
	}{
		{
			name:     "should expand ~ to home directory",
			path:     "~/.config/vigil/config.yaml",
			wantHome: true,
		},
		{
			name:     "should not modify absolute paths",
			path:     "/tmp/config.yaml",
			wantHome: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expanded := expandHomeDir(tt.path)

			if tt.wantHome {
				home, _ := os.UserHomeDir()
				if len(expanded) <= len(home) || expanded[:len(home)] != home {
					t.Errorf("expected path to start with home dir, got %s", expanded)
				}
			} else {
				if expanded != tt.path {
					t.Errorf("expected path unchanged, got %s", expanded)
				}
			}
		})
	}
}

func TestEnsureConfigDir(t *testing.T) {
	t.Run("should create config directory if it doesn't exist", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "deep", "nested", "config.yaml")

		err := ensureConfigDir(configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		dirPath := filepath.Dir(configPath)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Error("directory was not created")
		}
	})
}

func TestSaveConfig(t *testing.T) {
	t.Run("should save config to file", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")

		config := &Config{
			Services: []Service{
				{Name: "test-service", Ports: []int{8080}},
			},
		}

		err := SaveConfigToPath(config, configPath)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			t.Error("config file was not created")
		}

		// Verify we can load it back
		loaded, err := LoadConfigFromPath(configPath)
		if err != nil {
			t.Fatalf("failed to load saved config: %v", err)
		}

		if len(loaded.Services) != 1 {
			t.Errorf("expected 1 service, got %d", len(loaded.Services))
		}
	})
}

func TestLoadConfigPriority(t *testing.T) {
	t.Run("should prefer local config.yaml over default location", func(t *testing.T) {
		// Create a temporary directory and change to it
		tmpDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer os.Chdir(oldWd)

		os.Chdir(tmpDir)

		// Create a local config.yaml
		localConfig := &Config{
			Services: []Service{
				{Name: "local-service", Ports: []int{9999}},
			},
		}

		err := SaveConfigToPath(localConfig, "./config.yaml")
		if err != nil {
			t.Fatalf("failed to create local config: %v", err)
		}

		// Load config (should load local one)
		loaded, err := LoadConfig()
		if err != nil {
			t.Fatalf("failed to load config: %v", err)
		}

		if len(loaded.Services) == 0 {
			t.Fatal("expected at least one service")
		}

		if loaded.Services[0].Name != "local-service" {
			t.Errorf("expected local-service, got %s (local config was not prioritized)", loaded.Services[0].Name)
		}
	})
}
