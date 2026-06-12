package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Logger handles writing service output to log files and optionally to stdout.
type Logger struct {
	dir    string
	stdout bool
	mu     sync.Mutex
	files  map[string]*os.File
}

// New creates a new Logger that writes to the given directory.
// If stdout is true, output is also written to the terminal.
func New(dir string, stdout bool) *Logger {
	return &Logger{
		dir:    dir,
		stdout: stdout,
		files:  make(map[string]*os.File),
	}
}

// Writer returns an io.Writer that writes to both the log file and optionally stdout.
func (l *Logger) Writer(name string) (io.Writer, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if f, ok := l.files[name]; ok {
		return l.buildWriter(name, f), nil
	}

	if err := os.MkdirAll(l.dir, 0755); err != nil {
		return nil, fmt.Errorf("creating log dir: %w", err)
	}

	logPath := filepath.Join(l.dir, name+".log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening log file %s: %w", logPath, err)
	}

	l.files[name] = f
	return l.buildWriter(name, f), nil
}

// buildWriter creates a multi-writer for the given service.
func (l *Logger) buildWriter(name string, f *os.File) io.Writer {
	prefix := fmt.Sprintf("[%s] ", name)
	writers := []io.Writer{&prefixWriter{prefix: prefix, w: f}}
	if l.stdout {
		writers = append(writers, &prefixWriter{prefix: prefix, w: os.Stdout})
	}
	return io.MultiWriter(writers...)
}

// Close closes all open log files.
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var firstErr error
	for name, f := range l.files {
		if err := f.Close(); err != nil && firstErr == nil {
			firstErr = fmt.Errorf("closing log for %s: %w", name, err)
		}
		delete(l.files, name)
	}
	return firstErr
}

// LogPath returns the log file path for a service.
func (l *Logger) LogPath(name string) string {
	return filepath.Join(l.dir, name+".log")
}

// prefixWriter prepends a timestamp and service name prefix to each line.
type prefixWriter struct {
	prefix string
	w      io.Writer
}

func (pw *prefixWriter) Write(p []byte) (n int, err error) {
	// For simplicity, add prefix to the entire write.
	// In production, you'd want per-line prefixing.
	ts := time.Now().Format("15:04:05")
	header := fmt.Sprintf("%s %s", ts, pw.prefix)
	_, err = pw.w.Write(append([]byte(header), p...))
	return len(p), err
}
