package scriptgen

import (
	"fmt"
	"os"
	"path/filepath"
	"pitstop/internal/config"
	"text/template"
)

// Generator creates startup/shutdown scripts from a PitStop config.
type Generator struct {
	cfg       *config.Config
	outputDir string // where to write scripts
	logDir    string // where PID/log files go (relative to project root)
}

// New creates a Generator.
// cfg is the pitstop config, outputDir is where scripts are written,
// logDir is the log directory from config (used in scripts for PID/log paths).
func New(cfg *config.Config, outputDir string) *Generator {
	logDir := cfg.Logging.Dir
	if logDir == "" {
		logDir = "./logs"
	}
	return &Generator{
		cfg:       cfg,
		outputDir: outputDir,
		logDir:    logDir,
	}
}

// Generate creates all 4 scripts (start.sh, stop.sh, start.bat, stop.bat).
func (g *Generator) Generate() ([]string, error) {
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return nil, fmt.Errorf("creating output dir: %w", err)
	}

	data := g.buildData()

	files := []struct {
		name     string
		tpl      string
		perm     os.FileMode
	}{
		{"start.sh", startShTpl, 0755},
		{"stop.sh", stopShTpl, 0755},
		{"start.bat", startBatTpl, 0644},
		{"stop.bat", stopBatTpl, 0644},
	}

	var created []string
	for _, f := range files {
		path := filepath.Join(g.outputDir, f.name)
		if err := g.renderTemplate(path, f.tpl, data, f.perm); err != nil {
			return nil, fmt.Errorf("generating %s: %w", f.name, err)
		}
		created = append(created, path)
	}

	return created, nil
}

// buildData converts config into template data.
func (g *Generator) buildData() ScriptData {
	services := make([]ServiceData, 0, len(g.cfg.Services))
	for name, svc := range g.cfg.Services {
		services = append(services, ServiceData{
			Name:    name,
			Cwd:     svc.Path,
			Command: svc.Command,
		})
	}

	return ScriptData{
		ProjectName: g.cfg.Project.Name,
		Services:    services,
		OutputDir:   g.logDir,
	}
}

// renderTemplate renders a single template to a file.
func (g *Generator) renderTemplate(path, tplText string, data ScriptData, perm os.FileMode) error {
	tpl, err := template.New("").Parse(tplText)
	if err != nil {
		return fmt.Errorf("parsing template: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", path, err)
	}
	defer f.Close()

	if err := tpl.Execute(f, data); err != nil {
		return fmt.Errorf("rendering template: %w", err)
	}

	return nil
}
