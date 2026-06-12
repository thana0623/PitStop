package process

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ServiceProcess represents a running child process.
type ServiceProcess struct {
	Name    string
	Cmd     *exec.Cmd
	Started time.Time
}

// Manager handles starting, stopping, and tracking child processes.
type Manager struct {
	mu       sync.Mutex
	procs    map[string]*ServiceProcess
	pidDir   string
	shutdown chan struct{}
}

// New creates a new process manager.
// pidDir is the directory where PID files are stored.
func New(pidDir string) *Manager {
	return &Manager{
		procs:    make(map[string]*ServiceProcess),
		pidDir:   pidDir,
		shutdown: make(chan struct{}),
	}
}

// Start launches a child process for the given service.
// name is the service key from config, workDir is the working directory,
// and command is the shell command to execute.
func (m *Manager) Start(name, workDir, command string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.procs[name]; exists {
		return fmt.Errorf("service %q is already running", name)
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/C", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}
	cmd.Dir = workDir
	setProcessGroup(cmd)

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting service %q: %w", name, err)
	}

	m.procs[name] = &ServiceProcess{
		Name:    name,
		Cmd:     cmd,
		Started: time.Now(),
	}

	// Write PID file.
	if err := m.writePID(name, cmd.Process.Pid); err != nil {
		// Non-fatal: log but don't fail.
		fmt.Fprintf(os.Stderr, "warning: could not write PID file for %s: %v\n", name, err)
	}

	// Watch for process exit in background.
	go func() {
		err := cmd.Wait()
		m.mu.Lock()
		delete(m.procs, name)
		m.mu.Unlock()
		m.removePID(name)
		if err != nil {
			fmt.Fprintf(os.Stderr, "service %q exited: %v\n", name, err)
		} else {
			fmt.Printf("service %q exited normally\n", name)
		}
	}()

	return nil
}

// Stop gracefully stops a named service: SIGTERM, wait up to timeout, then SIGKILL.
func (m *Manager) Stop(name string, timeout time.Duration) error {
	m.mu.Lock()
	proc, exists := m.procs[name]
	m.mu.Unlock()

	if !exists {
		// Try reading PID file as fallback.
		return m.stopFromPIDFile(name, timeout)
	}

	return m.killProcess(proc.Cmd, timeout)
}

// StopAll gracefully stops all tracked services.
func (m *Manager) StopAll(timeout time.Duration) []error {
	m.mu.Lock()
	names := make([]string, 0, len(m.procs))
	for name := range m.procs {
		names = append(names, name)
	}
	m.mu.Unlock()

	var errs []error
	for _, name := range names {
		if err := m.Stop(name, timeout); err != nil {
			errs = append(errs, fmt.Errorf("stopping %s: %w", name, err))
		}
	}
	return errs
}

// StopFromPIDFiles stops all services found in PID files (used by `pitstop stop`).
func (m *Manager) StopFromPIDFiles(timeout time.Duration) []error {
	entries, err := os.ReadDir(m.pidDir)
	if err != nil {
		return []error{fmt.Errorf("reading PID dir: %w", err)}
	}

	var errs []error
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".pid") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".pid")
		if err := m.stopFromPIDFile(name, timeout); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// killProcess sends SIGTERM, waits, then SIGKILL if needed.
func (m *Manager) killProcess(cmd *exec.Cmd, timeout time.Duration) error {
	if cmd.Process == nil {
		return nil
	}

	pid := cmd.Process.Pid

	// Send SIGTERM (or TerminateProcess on Windows).
	if err := m.sendSignal(cmd); err != nil {
		// Process might already be dead — not an error.
		return nil
	}

	// Wait for the process to actually exit by polling the PID.
	// The background goroutine in Start() is also calling cmd.Wait(),
	// so we don't call Wait() here to avoid double-wait.
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !isProcessAlive(pid) {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	// Timeout — force kill.
	return m.forceKill(cmd)
}

// isProcessAlive checks if a process with the given PID is still alive.
func isProcessAlive(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		// On Windows, FindProcess always succeeds if the PID is valid.
		// Use tasklist to check if the process is actually running.
		out, err := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/NH").Output()
		if err != nil {
			return false
		}
		return len(out) > 0 && !strings.Contains(string(out), "No tasks")
	}
	// On Unix, send signal 0 to check if process exists.
	return proc.Signal(syscall.Signal(0)) == nil
}

// sendSignal sends SIGTERM on Unix or TerminateProcess on Windows.
func (m *Manager) sendSignal(cmd *exec.Cmd) error {
	if runtime.GOOS == "windows" {
		// On Windows, we use taskkill to send termination signal.
		out, err := exec.Command("taskkill", "/PID", strconv.Itoa(cmd.Process.Pid), "/T", "/F").CombinedOutput()
		if err != nil {
			return fmt.Errorf("taskkill: %w (output: %s)", err, string(out))
		}
		return nil
	}
	return cmd.Process.Signal(syscall.SIGTERM)
}

// forceKill sends SIGKILL or taskkill /F.
func (m *Manager) forceKill(cmd *exec.Cmd) error {
	if runtime.GOOS == "windows" {
		return exec.Command("taskkill", "/F", "/PID", strconv.Itoa(cmd.Process.Pid), "/T").Run()
	}
	return cmd.Process.Signal(syscall.SIGKILL)
}

// stopFromPIDFile reads a PID file and kills that process.
func (m *Manager) stopFromPIDFile(name string, timeout time.Duration) error {
	pidPath := filepath.Join(m.pidDir, name+".pid")
	data, err := os.ReadFile(pidPath)
	if err != nil {
		return fmt.Errorf("reading PID file for %s: %w", name, err)
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return fmt.Errorf("invalid PID in %s: %w", pidPath, err)
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("finding process %d: %w", pid, err)
	}

	// Send SIGTERM.
	if runtime.GOOS == "windows" {
		if err := exec.Command("taskkill", "/PID", strconv.Itoa(pid), "/T").Run(); err != nil {
			// Process might already be dead.
			m.removePID(name)
			return nil
		}
	} else {
		if err := proc.Signal(syscall.SIGTERM); err != nil {
			// Process might already be dead.
			m.removePID(name)
			return nil
		}
	}

	// Wait a bit, then force kill if needed.
	time.Sleep(timeout)
	if runtime.GOOS == "windows" {
		exec.Command("taskkill", "/F", "/PID", strconv.Itoa(pid), "/T").Run()
	} else {
		proc.Signal(syscall.SIGKILL)
	}

	m.removePID(name)
	return nil
}

// writePID writes the process PID to a file.
func (m *Manager) writePID(name string, pid int) error {
	if err := os.MkdirAll(m.pidDir, 0755); err != nil {
		return err
	}
	path := filepath.Join(m.pidDir, name+".pid")
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0644)
}

// removePID deletes the PID file for a service.
func (m *Manager) removePID(name string) {
	path := filepath.Join(m.pidDir, name+".pid")
	os.Remove(path) // ignore error
}

// IsRunning checks if a service is currently tracked as running.
func (m *Manager) IsRunning(name string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, exists := m.procs[name]
	return exists
}

// Running returns the names of all currently running services.
func (m *Manager) Running() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	names := make([]string, 0, len(m.procs))
	for name := range m.procs {
		names = append(names, name)
	}
	return names
}
