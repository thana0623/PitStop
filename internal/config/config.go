package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the pitstop.yaml configuration structure.
type Config struct {
	Project  Project            `yaml:"project"`
	Services map[string]Service `yaml:"services"`
	Logging  Logging            `yaml:"logging"`
}

// Project holds project-level settings.
type Project struct {
	Name string `yaml:"name"`
}

// Service defines a single service (backend, frontend, etc.).
type Service struct {
	Type        string `yaml:"type"`
	Path        string `yaml:"path"`
	Command     string `yaml:"command"`
	Port        int    `yaml:"port"`
	HealthCheck string `yaml:"health_check"`
}

// Logging holds logging configuration.
type Logging struct {
	Dir    string `yaml:"dir"`
	Stdout bool   `yaml:"stdout"`
}

// Load reads and parses a pitstop.yaml configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	cfg.applyDefaults()
	return cfg, nil
}

// validate checks required fields.
func (c *Config) validate() error {
	if c.Project.Name == "" {
		return fmt.Errorf("project.name is required")
	}
	if len(c.Services) == 0 {
		return fmt.Errorf("at least one service is required")
	}
	for name, svc := range c.Services {
		if svc.Command == "" {
			return fmt.Errorf("service %q: command is required", name)
		}
	}
	return nil
}

// applyDefaults sets default values for optional fields.
func (c *Config) applyDefaults() {
	if c.Logging.Dir == "" {
		c.Logging.Dir = "./logs"
	}
	for name, svc := range c.Services {
		if svc.Path == "" {
			svc.Path = "."
			c.Services[name] = svc
		}
	}
}
