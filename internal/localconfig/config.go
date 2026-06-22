package localconfig

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds machine-specific settings that don't travel with the project.
type Config struct {
	Paths map[string]string `yaml:"paths"`
	Ports PortsConfig        `yaml:"ports"`
}

// PortsConfig defines available port ranges on this machine.
type PortsConfig struct {
	Range [2]int `yaml:"range"`
}

// DefaultConfigDir returns ~/.pitstop/
func DefaultConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home dir: %w", err)
	}
	return filepath.Join(home, ".pitstop"), nil
}

// DefaultConfigPath returns ~/.pitstop/config.yaml
func DefaultConfigPath() (string, error) {
	dir, err := DefaultConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

// Load reads the local config from the given path.
// If path is empty, uses ~/.pitstop/config.yaml.
// If the file doesn't exists, returns a config with auto-detected paths.
func Load(path string) (*Config, error) {
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return nil, err
		}
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist — auto-detect everything.
			return AutoDetect()
		}
		return nil, fmt.Errorf("reading local config %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing local config %s: %w", path, err)
	}

	if cfg.Paths == nil {
		cfg.Paths = make(map[string]string)
	}

	// Fill in any missing paths with auto-detection.
	detected := DetectAll()
	for tool, path := range detected {
		if _, exists := cfg.Paths[tool]; !exists {
			cfg.Paths[tool] = path
		}
	}

	// Default port range.
	if cfg.Ports.Range == [2]int{0, 0} {
		cfg.Ports.Range = [2]int{10000, 60000}
	}

	return cfg, nil
}

// AutoDetect creates a config by detecting all tools from PATH.
func AutoDetect() (*Config, error) {
	cfg := &Config{
		Paths: DetectAll(),
		Ports: PortsConfig{Range: [2]int{10000, 60000}},
	}
	return cfg, nil
}

// Save writes the config to the given path (or default).
func (c *Config) Save(path string) error {
	if path == "" {
		var err error
		path, err = DefaultConfigPath()
		if err != nil {
			return err
		}
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config %s: %w", path, err)
	}

	return nil
}

// GetPath returns the resolved path for a tool, or the tool name itself if not configured.
func (c *Config) GetPath(tool string) string {
	if p, ok := c.Paths[tool]; ok && p != "" {
		return p
	}
	return tool // fallback: assume it's on PATH
}
