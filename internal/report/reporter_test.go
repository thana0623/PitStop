package report

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateNoLogs(t *testing.T) {
	dir := t.TempDir()
	r := New(dir)
	stats, err := r.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if len(stats) != 0 {
		t.Errorf("expected 0 stats, got %d", len(stats))
	}
}

func TestGenerateWithLogs(t *testing.T) {
	dir := t.TempDir()

	// Write a sample log file.
	logContent := `2024-01-01 INFO: Server started
2024-01-01 WARN: Port already in use
2024-01-01 ERROR: Connection failed
2024-01-01 INFO: Retrying
2024-01-01 FATAL: Cannot connect
`
	os.WriteFile(filepath.Join(dir, "backend.log"), []byte(logContent), 0644)

	r := New(dir)
	stats, err := r.Generate()
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(stats))
	}

	s := stats[0]
	if s.Name != "backend" {
		t.Errorf("Name = %q, want %q", s.Name, "backend")
	}
	if s.Total != 5 {
		t.Errorf("Total = %d, want 5", s.Total)
	}
	if s.Error != 2 { // ERROR + FATAL
		t.Errorf("Error = %d, want 2", s.Error)
	}
	if s.Warn != 1 {
		t.Errorf("Warn = %d, want 1", s.Warn)
	}
	if s.Info != 2 {
		t.Errorf("Info = %d, want 2", s.Info)
	}
}

func TestFormatEmpty(t *testing.T) {
	out := Format(nil)
	if out == "" {
		t.Error("Format should not return empty string")
	}
}

func TestFormatWithStats(t *testing.T) {
	stats := []Stats{
		{Name: "backend", Total: 100, Error: 3, Warn: 5, Info: 90},
		{Name: "frontend", Total: 50, Error: 0, Warn: 1, Info: 49},
	}
	out := Format(stats)
	if out == "" {
		t.Error("Format should not return empty string")
	}
}
