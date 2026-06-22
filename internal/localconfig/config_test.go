package localconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NonExistentFile_AutoDetects(t *testing.T) {
	cfg, err := Load("/nonexistent/path/config.yaml")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.Paths == nil {
		t.Fatal("expected Paths to be non-nil")
	}
	if cfg.Ports.Range != [2]int{10000, 60000} {
		t.Errorf("expected default port range [10000, 60000], got %v", cfg.Ports.Range)
	}
}

func TestLoad_ValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := `
paths:
  java: /usr/bin/java
  node: /usr/local/bin/node
ports:
  range: [20000, 50000]
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Paths["java"] != "/usr/bin/java" {
		t.Errorf("expected java=/usr/bin/java, got %s", cfg.Paths["java"])
	}
	if cfg.Paths["node"] != "/usr/local/bin/node" {
		t.Errorf("expected node=/usr/local/bin/node, got %s", cfg.Paths["node"])
	}
	if cfg.Ports.Range != [2]int{20000, 50000} {
		t.Errorf("expected port range [20000, 50000], got %v", cfg.Ports.Range)
	}
}

func TestSave_And_Load_Roundtrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	original := &Config{
		Paths: map[string]string{
			"java": "/opt/java/bin/java",
		},
		Ports: PortsConfig{Range: [2]int{30000, 40000}},
	}

	if err := original.Save(path); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load failed: %v", err)
	}

	if loaded.Paths["java"] != "/opt/java/bin/java" {
		t.Errorf("expected java=/opt/java/bin/java, got %s", loaded.Paths["java"])
	}
	if loaded.Ports.Range != [2]int{30000, 40000} {
		t.Errorf("expected port range [30000, 40000], got %v", loaded.Ports.Range)
	}
}

func TestGetPath_Configured(t *testing.T) {
	cfg := &Config{Paths: map[string]string{"java": "/usr/bin/java"}}
	if got := cfg.GetPath("java"); got != "/usr/bin/java" {
		t.Errorf("expected /usr/bin/java, got %s", got)
	}
}

func TestGetPath_NotConfigured_Fallback(t *testing.T) {
	cfg := &Config{Paths: map[string]string{}}
	if got := cfg.GetPath("java"); got != "java" {
		t.Errorf("expected 'java', got %s", got)
	}
}

func TestDetectPath_FindsKnownTool(t *testing.T) {
	// git should be available on most dev machines.
	p := DetectPath("git")
	if p == "" {
		t.Skip("git not found on PATH, skipping")
	}
	if !filepath.IsAbs(p) {
		t.Errorf("expected absolute path, got %s", p)
	}
}

func TestDetectPath_UnknownTool(t *testing.T) {
	p := DetectPath("this_tool_definitely_does_not_exist_12345")
	if p != "" {
		t.Errorf("expected empty string for unknown tool, got %s", p)
	}
}
