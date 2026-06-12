package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadValidConfig(t *testing.T) {
	content := `
project:
  name: "test-app"
services:
  backend:
    type: springboot
    path: "./backend"
    command: "mvn spring-boot:run"
    port: 8080
    health_check: "http://localhost:8080/health"
  frontend:
    type: vue
    path: "./frontend"
    command: "pnpm dev"
    port: 3000
logging:
  dir: "./logs"
  stdout: true
`
	path := writeTempFile(t, "pitstop.yaml", content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Project.Name != "test-app" {
		t.Errorf("Project.Name = %q, want %q", cfg.Project.Name, "test-app")
	}
	if len(cfg.Services) != 2 {
		t.Errorf("len(Services) = %d, want 2", len(cfg.Services))
	}
	backend := cfg.Services["backend"]
	if backend.Type != "springboot" {
		t.Errorf("backend.Type = %q, want %q", backend.Type, "springboot")
	}
	if backend.Port != 8080 {
		t.Errorf("backend.Port = %d, want 8080", backend.Port)
	}
	if cfg.Logging.Dir != "./logs" {
		t.Errorf("Logging.Dir = %q, want %q", cfg.Logging.Dir, "./logs")
	}
	if !cfg.Logging.Stdout {
		t.Error("Logging.Stdout should be true")
	}
}

func TestLoadMissingProjectName(t *testing.T) {
	content := `
project:
  name: ""
services:
  backend:
    command: "echo hello"
`
	path := writeTempFile(t, "pitstop.yaml", content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() should fail when project.name is empty")
	}
}

func TestLoadNoServices(t *testing.T) {
	content := `
project:
  name: "test"
services: {}
`
	path := writeTempFile(t, "pitstop.yaml", content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() should fail when no services defined")
	}
}

func TestLoadMissingCommand(t *testing.T) {
	content := `
project:
  name: "test"
services:
  backend:
    type: springboot
    path: "./backend"
`
	path := writeTempFile(t, "pitstop.yaml", content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() should fail when service command is empty")
	}
}

func TestLoadDefaultValues(t *testing.T) {
	content := `
project:
  name: "test"
services:
  app:
    command: "echo hello"
logging:
  dir: ""
`
	path := writeTempFile(t, "pitstop.yaml", content)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.Logging.Dir != "./logs" {
		t.Errorf("default Logging.Dir = %q, want %q", cfg.Logging.Dir, "./logs")
	}
	if cfg.Services["app"].Path != "." {
		t.Errorf("default Service.Path = %q, want %q", cfg.Services["app"].Path, ".")
	}
}

func TestLoadFileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/pitstop.yaml")
	if err == nil {
		t.Fatal("Load() should fail when file does not exist")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	content := `{{{invalid yaml`
	path := writeTempFile(t, "pitstop.yaml", content)
	_, err := Load(path)
	if err == nil {
		t.Fatal("Load() should fail on invalid YAML")
	}
}

func TestLoadActualConfig(t *testing.T) {
	// Test with the actual pitstop.yaml in the project root
	cfgPath := filepath.Join("..", "..", "pitstop.yaml")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		t.Skip("pitstop.yaml not found, skipping")
	}
	cfg, err := Load(cfgPath)
	if err != nil {
		t.Fatalf("Load(%s) error: %v", cfgPath, err)
	}
	if cfg.Project.Name == "" {
		t.Error("project.name should not be empty")
	}
	if len(cfg.Services) == 0 {
		t.Error("should have at least one service")
	}
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("writeTempFile: %v", err)
	}
	return path
}
