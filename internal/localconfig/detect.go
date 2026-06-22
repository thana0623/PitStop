package localconfig

import (
	"os/exec"
	"runtime"
	"strings"
)

// wellKnownTools is the list of tools PitStop knows how to detect.
var wellKnownTools = []string{
	"java", "node", "npm", "pnpm", "yarn",
	"mvn", "gradle",
	"python", "python3", "pip",
	"mysql", "redis-cli",
	"docker", "git",
}

// DetectAll finds all well-known tools on PATH.
// Returns a map of tool name → absolute path.
func DetectAll() map[string]string {
	result := make(map[string]string)
	for _, tool := range wellKnownTools {
		if p := DetectPath(tool); p != "" {
			result[tool] = p
		}
	}
	return result
}

// DetectPath finds the absolute path of a single tool.
// Returns "" if not found.
func DetectPath(tool string) string {
	// On Windows, try common extensions.
	if runtime.GOOS == "windows" {
		for _, ext := range []string{"", ".exe", ".cmd", ".bat"} {
			if p, err := exec.LookPath(tool + ext); err == nil {
				return p
			}
		}
		return ""
	}

	p, err := exec.LookPath(tool)
	if err != nil {
		return ""
	}
	return p
}

// DetectVersion tries to get the version of a tool by running common flags.
// Returns the first line of output that looks like a version, or "".
func DetectVersion(tool string, path string) string {
	if path == "" {
		path = tool
	}

	// Common version flags per tool.
	flags := versionFlags(tool)

	for _, flag := range flags {
		out, err := exec.Command(path, flag...).CombinedOutput()
		if err != nil {
			continue
		}
		line := firstNonEmptyLine(string(out))
		if line != "" {
			return line
		}
	}
	return ""
}

// versionFlags returns the common version flags for a tool.
func versionFlags(tool string) [][]string {
	base := strings.ToLower(tool)

	switch {
	case base == "java" || base == "javac":
		return [][]string{{"--version"}, {"-version"}}
	case base == "node":
		return [][]string{{"--version"}, {"-v"}}
	case base == "npm" || base == "pnpm" || base == "yarn":
		return [][]string{{"--version"}, {"-v"}}
	case base == "mvn" || base == "mvnw":
		return [][]string{{"--version"}, {"-version"}}
	case base == "gradle" || base == "gradlew":
		return [][]string{{"--version"}}
	case base == "python" || base == "python3":
		return [][]string{{"--version"}, {"-V"}}
	case base == "pip" || base == "pip3":
		return [][]string{{"--version"}}
	case base == "mysql":
		return [][]string{{"--version"}}
	case base == "redis-cli":
		return [][]string{{"--version"}}
	case base == "docker":
		return [][]string{{"--version"}}
	case base == "git":
		return [][]string{{"--version"}}
	default:
		return [][]string{{"--version"}, {"-version"}, {"-v"}, {"version"}}
	}
}

// firstNonEmptyLine returns the first non-empty trimmed line from text.
func firstNonEmptyLine(text string) string {
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			return trimmed
		}
	}
	return ""
}
