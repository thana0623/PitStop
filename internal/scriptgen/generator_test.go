package scriptgen

import (
	"os"
	"path/filepath"
	"pitstop/internal/config"
	"runtime"
	"strings"
	"testing"
)

func TestGenerate_CreatesFourFiles(t *testing.T) {
	cfg := &config.Config{
		Project: config.Project{Name: "test-app"},
		Services: map[string]config.Service{
			"backend": {
				Path:    "./backend",
				Command: "mvn spring-boot:run",
			},
			"frontend": {
				Path:    "./frontend",
				Command: "npm run dev",
			},
		},
		Logging: config.Logging{Dir: "./logs"},
	}

	outDir := t.TempDir()
	gen := New(cfg, outDir)

	files, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	if len(files) != 4 {
		t.Fatalf("expected 4 files, got %d", len(files))
	}

	expectedNames := []string{"start.sh", "stop.sh", "start.bat", "stop.bat"}
	for _, name := range expectedNames {
		path := filepath.Join(outDir, name)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file %s to exist", name)
		}
	}
}

func TestGenerate_StartShContent(t *testing.T) {
	cfg := &config.Config{
		Project: config.Project{Name: "my-project"},
		Services: map[string]config.Service{
			"api": {
				Path:    "./api",
				Command: "go run main.go",
			},
		},
		Logging: config.Logging{Dir: "./logs"},
	}

	outDir := t.TempDir()
	gen := New(cfg, outDir)
	_, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "start.sh"))
	if err != nil {
		t.Fatalf("reading start.sh: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "my-project") {
		t.Error("start.sh missing project name")
	}
	if !strings.Contains(s, "go run main.go") {
		t.Error("start.sh missing command")
	}
	if !strings.Contains(s, "./api") {
		t.Error("start.sh missing cwd")
	}
	if !strings.Contains(s, "api.pid") {
		t.Error("start.sh missing PID file reference")
	}
}

func TestGenerate_StopShContent(t *testing.T) {
	cfg := &config.Config{
		Project: config.Project{Name: "stop-test"},
		Services: map[string]config.Service{
			"svc": {Path: ".", Command: "echo hello"},
		},
		Logging: config.Logging{Dir: "./logs"},
	}

	outDir := t.TempDir()
	gen := New(cfg, outDir)
	_, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "stop.sh"))
	if err != nil {
		t.Fatalf("reading stop.sh: %v", err)
	}

	s := string(content)
	if !strings.Contains(s, "stop-test") {
		t.Error("stop.sh missing project name")
	}
	if !strings.Contains(s, ".pid") {
		t.Error("stop.sh missing PID file handling")
	}
}

func TestGenerate_ShScriptsAreExecutable(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix file permissions not applicable on Windows")
	}

	cfg := &config.Config{
		Project: config.Project{Name: "perm-test"},
		Services: map[string]config.Service{
			"svc": {Path: ".", Command: "echo"},
		},
		Logging: config.Logging{Dir: "./logs"},
	}

	outDir := t.TempDir()
	gen := New(cfg, outDir)
	_, err := gen.Generate()
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	for _, name := range []string{"start.sh", "stop.sh"} {
		info, err := os.Stat(filepath.Join(outDir, name))
		if err != nil {
			t.Fatalf("stat %s: %v", name, err)
		}
		if info.Mode().Perm()&0100 == 0 {
			t.Errorf("%s should be executable, got %v", name, info.Mode())
		}
	}
}
