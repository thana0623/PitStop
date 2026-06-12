package report

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// Stats holds log statistics for a single service.
type Stats struct {
	Name  string
	Total int
	Error int
	Warn  int
	Info  int
}

// Reporter generates log summary reports.
type Reporter struct {
	logDir string
}

// New creates a new Reporter for the given log directory.
func New(logDir string) *Reporter {
	return &Reporter{logDir: logDir}
}

// Generate scans all .log files and returns stats per service.
func (r *Reporter) Generate() ([]Stats, error) {
	entries, err := os.ReadDir(r.logDir)
	if err != nil {
		return nil, fmt.Errorf("reading log dir: %w", err)
	}

	var allStats []Stats
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".log")
		path := filepath.Join(r.logDir, entry.Name())
		stats, err := r.analyze(name, path)
		if err != nil {
			return nil, fmt.Errorf("analyzing %s: %w", path, err)
		}
		allStats = append(allStats, stats)
	}

	return allStats, nil
}

var (
	errorRe = regexp.MustCompile(`(?i)\berror\b|\bfatal\b|\bpanic\b`)
	warnRe  = regexp.MustCompile(`(?i)\bwarn\b|\bwarning\b`)
	infoRe  = regexp.MustCompile(`(?i)\binfo\b`)
)

// analyze reads a log file and counts log levels.
func (r *Reporter) analyze(name, path string) (Stats, error) {
	f, err := os.Open(path)
	if err != nil {
		return Stats{}, err
	}
	defer f.Close()

	stats := Stats{Name: name}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		stats.Total++
		switch {
		case errorRe.MatchString(line):
			stats.Error++
		case warnRe.MatchString(line):
			stats.Warn++
		case infoRe.MatchString(line):
			stats.Info++
		}
	}

	return stats, scanner.Err()
}

// Format prints the report in a human-readable format.
func Format(stats []Stats) string {
	if len(stats) == 0 {
		return "No log files found."
	}

	var b strings.Builder
	b.WriteString("=== PitStop Log Report ===\n\n")

	totalErrors := 0
	totalWarns := 0
	totalLines := 0

	for _, s := range stats {
		totalErrors += s.Error
		totalWarns += s.Warn
		totalLines += s.Total

		b.WriteString(fmt.Sprintf("  %-12s  lines: %d  errors: %d  warnings: %d  info: %d\n",
			s.Name, s.Total, s.Error, s.Warn, s.Info))
	}

	b.WriteString(fmt.Sprintf("\n  Total: %d lines, %d errors, %d warnings\n", totalLines, totalErrors, totalWarns))

	if totalErrors > 0 {
		b.WriteString("\n  ⚠️  Errors found — check logs for details.\n")
	}

	return b.String()
}
