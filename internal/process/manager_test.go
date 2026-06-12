package process

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

// longRunningCmd returns a command that runs for a long time on the current platform.
func longRunningCmd() string {
	if runtime.GOOS == "windows" {
		// Use ping as a sleep alternative that works in non-interactive contexts.
		return "ping -n 61 127.0.0.1"
	}
	return "sleep 60"
}

func TestStartAndStop(t *testing.T) {
	pidDir := t.TempDir()
	mgr := New(pidDir)

	if err := mgr.Start("test-svc", ".", longRunningCmd()); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if !mgr.IsRunning("test-svc") {
		t.Error("service should be running after Start()")
	}

	// PID file should exist.
	pidPath := filepath.Join(pidDir, "test-svc.pid")
	if _, err := os.Stat(pidPath); os.IsNotExist(err) {
		t.Error("PID file should exist after Start()")
	}

	// Stop the service.
	if err := mgr.Stop("test-svc", 5*time.Second); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}

	// Wait for background goroutine to clean up.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !mgr.IsRunning("test-svc") {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if mgr.IsRunning("test-svc") {
		t.Error("service should not be running after Stop()")
	}
}

func TestStartDuplicate(t *testing.T) {
	pidDir := t.TempDir()
	mgr := New(pidDir)

	if err := mgr.Start("dup", ".", longRunningCmd()); err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer mgr.Stop("dup", 2*time.Second)

	// Starting the same service again should fail.
	if err := mgr.Start("dup", ".", longRunningCmd()); err == nil {
		t.Error("Start() should fail for duplicate service")
	}
}

func TestStopNotRunning(t *testing.T) {
	pidDir := t.TempDir()
	mgr := New(pidDir)

	// Stopping a non-existent service should return an error (no PID file).
	err := mgr.Stop("ghost", 2*time.Second)
	if err == nil {
		t.Error("Stop() should error for non-existent service")
	}
}

func TestStopAll(t *testing.T) {
	pidDir := t.TempDir()
	mgr := New(pidDir)

	mgr.Start("svc1", ".", longRunningCmd())
	mgr.Start("svc2", ".", longRunningCmd())
	time.Sleep(200 * time.Millisecond)

	errs := mgr.StopAll(5 * time.Second)
	if len(errs) != 0 {
		t.Errorf("StopAll() errors: %v", errs)
	}

	// Wait for background goroutines to clean up.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if len(mgr.Running()) == 0 {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if len(mgr.Running()) != 0 {
		t.Errorf("expected 0 running, got %d", len(mgr.Running()))
	}
}

func TestRunning(t *testing.T) {
	pidDir := t.TempDir()
	mgr := New(pidDir)

	if len(mgr.Running()) != 0 {
		t.Error("new manager should have 0 running services")
	}

	mgr.Start("a", ".", longRunningCmd())
	mgr.Start("b", ".", longRunningCmd())
	defer mgr.StopAll(2 * time.Second)

	names := mgr.Running()
	if len(names) != 2 {
		t.Errorf("expected 2 running, got %d", len(names))
	}
}

func TestProcessExit(t *testing.T) {
	pidDir := t.TempDir()
	mgr := New(pidDir)

	// Start a process that exits quickly.
	if err := mgr.Start("quick", ".", "echo done"); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	// Wait for the process to exit.
	time.Sleep(500 * time.Millisecond)

	if mgr.IsRunning("quick") {
		t.Error("service should not be running after exit")
	}

	// PID file should be cleaned up.
	pidPath := filepath.Join(pidDir, "quick.pid")
	if _, err := os.Stat(pidPath); !os.IsNotExist(err) {
		t.Error("PID file should be removed after process exit")
	}
}
