package initcmd

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// ServiceInfo holds detected service information.
type ServiceInfo struct {
	Name        string
	Type        string
	Path        string
	Command     string
	Port        int
	HealthCheck string
}

// ScanResult holds the full scan output.
type ScanResult struct {
	ProjectName  string
	Services     []ServiceInfo
	Requirements map[string]string
}

// directories to skip when scanning for services.
var skipDirs = map[string]bool{
	".git":         true,
	".claude":      true,
	".idea":        true,
	".vscode":      true,
	"node_modules": true,
	"__pycache__":  true,
	".venv":        true,
	"venv":         true,
	"dist":         true,
	"build":        true,
	"target":       true,
	"docs":         true,
	"doc":          true,
	"scripts":      true,
	"docker":       true,
	".docker":      true,
	"logs":         true,
	"tmp":          true,
	"temp":         true,
}

// Scan scans the given directory and returns detected services.
func Scan(dir string) (*ScanResult, error) {
	result := &ScanResult{
		ProjectName:  filepath.Base(dir),
		Requirements: make(map[string]string),
	}

	// First, check if the root directory itself is a service.
	if svc := detectService(dir, ""); svc != nil {
		result.Services = append(result.Services, *svc)
	}

	// Then scan immediate subdirectories for additional services.
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if skipDirs[name] || strings.HasPrefix(name, ".") {
			continue
		}

		subdir := filepath.Join(dir, name)
		if svc := detectService(subdir, name); svc != nil {
			result.Services = append(result.Services, *svc)
		}
	}

	// Infer requirements from detected services.
	for _, svc := range result.Services {
		switch svc.Type {
		case "springboot":
			if ver := detectJavaVersion(svc.Path); ver != "" {
				result.Requirements["java"] = ">=" + ver
			}
			result.Requirements["maven"] = ">=3.8"
		case "vue", "react", "angular", "node":
			if ver := detectNodeVersion(svc.Path); ver != "" {
				result.Requirements["node"] = ">=" + ver
			}
			if _, err := os.Stat(filepath.Join(svc.Path, "pnpm-lock.yaml")); err == nil {
				result.Requirements["pnpm"] = ">=9"
			}
		}
	}

	return result, nil
}

// detectService checks if a directory contains a recognizable project.
func detectService(dir, overrideName string) *ServiceInfo {
	// Spring Boot (Maven)
	if isFile(filepath.Join(dir, "pom.xml")) {
		if hasSpringBoot(filepath.Join(dir, "pom.xml")) {
			name := overrideName
			if name == "" {
				name = "backend"
			}
			port := detectSpringBootPort(dir)
			if port == 0 {
				port = 8080
			}
			return &ServiceInfo{
				Name:        name,
				Type:        "springboot",
				Path:        dir,
				Command:     "mvn spring-boot:run",
				Port:        port,
				HealthCheck: fmt.Sprintf("http://localhost:%d/actuator/health", port),
			}
		}
	}

	// Spring Boot (Gradle)
	if isFile(filepath.Join(dir, "build.gradle")) || isFile(filepath.Join(dir, "build.gradle.kts")) {
		name := overrideName
		if name == "" {
			name = "backend"
		}
		port := detectSpringBootPort(dir)
		if port == 0 {
			port = 8080
		}
		return &ServiceInfo{
			Name:        name,
			Type:        "springboot",
			Path:        dir,
			Command:     "gradle bootRun",
			Port:        port,
			HealthCheck: fmt.Sprintf("http://localhost:%d/actuator/health", port),
		}
	}

	// Node.js projects
	if isFile(filepath.Join(dir, "package.json")) {
		pkg := readPackageJSON(filepath.Join(dir, "package.json"))
		if pkg == nil {
			return nil
		}

		// Skip workspace root package.json (monorepo manager, not a service).
		if pkg.Workspaces != nil {
			return nil
		}
		// Skip package.json with no name and no scripts (likely a workspace root).
		if pkg.Name == "" && len(pkg.Scripts) == 0 {
			return nil
		}

		svcType, command := detectNodeProjectType(dir, pkg)
		name := overrideName
		if name == "" {
			name = svcType
		}

		port := detectNodePort(dir, pkg)
		if port == 0 {
			port = defaultPort(svcType)
		}

		healthCheck := fmt.Sprintf("http://localhost:%d", port)

		return &ServiceInfo{
			Name:        name,
			Type:        svcType,
			Path:        dir,
			Command:     command,
			Port:        port,
			HealthCheck: healthCheck,
		}
	}

	// Go
	if isFile(filepath.Join(dir, "go.mod")) {
		name := overrideName
		if name == "" {
			name = "backend"
		}
		return &ServiceInfo{
			Name:    name,
			Type:    "go",
			Path:    dir,
			Command: "go run .",
			Port:    8080,
		}
	}

	// Python (Django)
	if isFile(filepath.Join(dir, "manage.py")) {
		name := overrideName
		if name == "" {
			name = "backend"
		}
		return &ServiceInfo{
			Name:    name,
			Type:    "django",
			Path:    dir,
			Command: "python manage.py runserver",
			Port:    8000,
		}
	}

	return nil
}

// packageJSON is a minimal representation for reading package.json.
type packageJSON struct {
	Name         string            `json:"name"`
	Scripts      map[string]string `json:"scripts"`
	Dependencies map[string]string `json:"dependencies"`
	Engines      map[string]string `json:"engines"`
	Workspaces   interface{}       `json:"workspaces"` // string or []string
}

func readPackageJSON(path string) *packageJSON {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var pkg packageJSON
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil
	}
	return &pkg
}

// detectNodeProjectType determines the framework and dev command.
func detectNodeProjectType(dir string, pkg *packageJSON) (svcType, command string) {
	deps := pkg.Dependencies
	if deps == nil {
		deps = make(map[string]string)
	}

	// Check for Vue
	if _, ok := deps["vue"]; ok {
		return "vue", findDevCommand(pkg)
	}
	// Check for React
	if _, ok := deps["react"]; ok {
		return "react", findDevCommand(pkg)
	}
	// Check for Angular
	if _, ok := deps["@angular/core"]; ok {
		return "angular", "ng serve"
	}

	return "node", findDevCommand(pkg)
}

// findDevCommand extracts the dev/start script from package.json.
func findDevCommand(pkg *packageJSON) string {
	if pkg.Scripts == nil {
		return "pnpm dev"
	}
	// Prefer "dev", then "start".
	if _, ok := pkg.Scripts["dev"]; ok {
		return "pnpm dev"
	}
	if _, ok := pkg.Scripts["start"]; ok {
		return "pnpm start"
	}
	return "pnpm dev"
}

// detectNodePort tries to find the port from vite.config or package.json scripts.
func detectNodePort(dir string, pkg *packageJSON) int {
	// Check vite.config.ts / vite.config.js
	for _, name := range []string{"vite.config.ts", "vite.config.js"} {
		if port := extractVitePort(filepath.Join(dir, name)); port > 0 {
			return port
		}
	}
	// Check angular.json
	if port := extractAngularPort(filepath.Join(dir, "angular.json")); port > 0 {
		return port
	}
	return 0
}

// extractVitePort reads a vite config and extracts server.port.
func extractVitePort(path string) int {
	if !isFile(path) {
		return 0
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	// Look for port: XXXX in server config.
	re := regexp.MustCompile(`port\s*:\s*(\d+)`)
	matches := re.FindSubmatch(data)
	if len(matches) >= 2 {
		if port, err := strconv.Atoi(string(matches[1])); err == nil {
			return port
		}
	}
	return 0
}

// extractAngularPort reads angular.json and extracts serve port.
func extractAngularPort(path string) int {
	if !isFile(path) {
		return 0
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	re := regexp.MustCompile(`"port"\s*:\s*(\d+)`)
	matches := re.FindSubmatch(data)
	if len(matches) >= 2 {
		if port, err := strconv.Atoi(string(matches[1])); err == nil {
			return port
		}
	}
	return 0
}

// defaultPort returns the default port for a project type.
func defaultPort(svcType string) int {
	switch svcType {
	case "springboot":
		return 8080
	case "vue", "react":
		return 5173
	case "angular":
		return 4200
	case "django":
		return 8000
	case "go":
		return 8080
	default:
		return 3000
	}
}

// hasSpringBoot checks if a pom.xml contains spring-boot dependencies.
func hasSpringBoot(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, "spring-boot") ||
		strings.Contains(content, "org.springframework.boot")
}

// detectSpringBootPort reads application.yml/properties for server.port.
func detectSpringBootPort(dir string) int {
	for _, name := range []string{
		"src/main/resources/application.yml",
		"src/main/resources/application.yaml",
		"src/main/resources/application.properties",
	} {
		path := filepath.Join(dir, name)
		if !isFile(path) {
			continue
		}
		if port := extractServerPort(path); port > 0 {
			return port
		}
	}
	return 0
}

// extractServerPort reads server.port from a YAML or properties file.
func extractServerPort(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	portRe := regexp.MustCompile(`server\.port\s*[=:]\s*(.+)`)
	yamlRe := regexp.MustCompile(`port\s*:\s*(.+)`)

	inServerBlock := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Properties format: server.port=8080 or server.port=${SERVER_PORT:9001}
		if m := portRe.FindStringSubmatch(line); len(m) >= 2 {
			val := strings.TrimSpace(m[1])
			if p := extractSpringDefault(val); p > 0 {
				return p
			}
			if p, err := strconv.Atoi(val); err == nil {
				return p
			}
		}

		// YAML format: server:\n  port: 8080
		if strings.HasPrefix(line, "server:") {
			inServerBlock = true
			continue
		}
		if inServerBlock && !strings.HasPrefix(line, "#") {
			if m := yamlRe.FindStringSubmatch(line); len(m) >= 2 {
				val := strings.TrimSpace(m[1])
				val = strings.Trim(val, "\"'")
				// Handle Spring placeholder: ${SERVER_PORT:9001}
				if p := extractSpringDefault(val); p > 0 {
					return p
				}
				if p, err := strconv.Atoi(val); err == nil {
					return p
				}
			}
			// Exit server block if we hit a non-indented line.
			if len(line) > 0 && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
				inServerBlock = false
			}
		}
	}
	return 0
}

// detectJavaVersion reads java.version from pom.xml.
func detectJavaVersion(dir string) string {
	path := filepath.Join(dir, "pom.xml")
	if !isFile(path) {
		return ""
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	// <java.version>17</java.version>
	re := regexp.MustCompile(`<java\.version>(\d+)</java\.version>`)
	if m := re.FindSubmatch(data); len(m) >= 2 {
		return string(m[1])
	}
	// <maven.compiler.source>17</maven.compiler.source>
	re = regexp.MustCompile(`<maven\.compiler\.source>(\d+)</maven\.compiler\.source>`)
	if m := re.FindSubmatch(data); len(m) >= 2 {
		return string(m[1])
	}
	return ""
}

// detectNodeVersion reads node version from package.json engines or .nvmrc.
func detectNodeVersion(dir string) string {
	// Check package.json engines.node
	pkg := readPackageJSON(filepath.Join(dir, "package.json"))
	if pkg != nil {
		if ver, ok := pkg.Engines["node"]; ok {
			return cleanVersion(ver)
		}
	}
	// Check .nvmrc
	nvmrc := filepath.Join(dir, ".nvmrc")
	if isFile(nvmrc) {
		data, err := os.ReadFile(nvmrc)
		if err == nil {
			return cleanVersion(strings.TrimSpace(string(data)))
		}
	}
	return ""
}

// cleanVersion extracts a simple semver from complex version strings like ">=20.19.0 || >=22.12.0".
func cleanVersion(ver string) string {
	// Take the first part before || or space.
	ver = strings.Split(ver, "||")[0]
	ver = strings.TrimSpace(ver)
	// Strip leading operators.
	ver = strings.TrimLeft(ver, ">=<~^")
	// Keep only major.minor.patch.
	parts := strings.SplitN(ver, ".", 3)
	if len(parts) == 0 {
		return ""
	}
	// Return major only if that's all we have.
	if len(parts) == 1 {
		return parts[0]
	}
	return parts[0] + "." + parts[1]
}

// extractSpringDefault parses Spring placeholder ${VAR:default} and returns the default value as int.
func extractSpringDefault(val string) int {
	// Match ${VAR:default} or ${VAR:default_value}
	if !strings.HasPrefix(val, "${") || !strings.HasSuffix(val, "}") {
		return 0
	}
	inner := val[2 : len(val)-1]
	parts := strings.SplitN(inner, ":", 2)
	if len(parts) < 2 {
		return 0
	}
	p, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	return p
}

// isFile checks if a path exists and is a regular file.
func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
