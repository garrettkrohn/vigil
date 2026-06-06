package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Service represents a microservice with its name and port numbers
type Service struct {
	Name   string `yaml:"name"`
	Ports  []int  `yaml:"ports"`
	Docker bool   `yaml:"docker,omitempty"`
}

// Config represents the vigil configuration
type Config struct {
	Services []Service `yaml:"services"`
}

// GetDefaultConfigPath returns the default config file path
func GetDefaultConfigPath() string {
	return expandHomeDir("~/.config/vigil/config.yaml")
}

// LoadConfig loads the configuration, checking current directory first
func LoadConfig() (*Config, error) {
	// First, check for config.yaml in current directory
	localConfig := "./config.yaml"
	if _, err := os.Stat(localConfig); err == nil {
		return LoadConfigFromPath(localConfig)
	}

	// Fall back to default location
	return LoadConfigFromPath(GetDefaultConfigPath())
}

// LoadConfigFromPath loads the configuration from a specific path
func LoadConfigFromPath(path string) (*Config, error) {
	expandedPath := expandHomeDir(path)

	// If file doesn't exist, create default config
	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		return createDefaultConfig(), nil
	}

	data, err := os.ReadFile(expandedPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	return &config, nil
}

// SaveConfig saves the configuration to the default location
func SaveConfig(config *Config) error {
	return SaveConfigToPath(config, GetDefaultConfigPath())
}

// SaveConfigToPath saves the configuration to a specific path
func SaveConfigToPath(config *Config, path string) error {
	expandedPath := expandHomeDir(path)

	if err := ensureConfigDir(expandedPath); err != nil {
		return err
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(expandedPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// EnsureConfigExists creates the default config file if it doesn't exist
func EnsureConfigExists() error {
	configPath := GetDefaultConfigPath()
	expandedPath := expandHomeDir(configPath)

	if _, err := os.Stat(expandedPath); os.IsNotExist(err) {
		defaultConfig := createDefaultConfig()
		return SaveConfig(defaultConfig)
	}

	return nil
}

// expandHomeDir expands ~ to the user's home directory
func expandHomeDir(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}

	if path == "~" {
		return home
	}

	return filepath.Join(home, path[2:])
}

// ensureConfigDir creates the directory for the config file if it doesn't exist
func ensureConfigDir(path string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	return nil
}

// createDefaultConfig creates a default configuration with example services
func createDefaultConfig() *Config {
	return &Config{
		Services: []Service{
			{
				Name:  "example-api",
				Ports: []int{8080},
			},
			{
				Name:  "example-auth",
				Ports: []int{8081, 9091},
			},
		},
	}
}
