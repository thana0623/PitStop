package envcheck

import (
	"fmt"
	"os/exec"
	"pitstop/internal/localconfig"
	"runtime"
	"strings"
)

// CheckResult holds the result of a single environment check.
type CheckResult struct {
	Tool       string // e.g., "java"
	Required   string // e.g., ">=17"
	Found      bool   // whether the tool was found
	ActualPath string // resolved path
	ActualVer  string // detected version string
	ParsedVer  string // extracted version number
	Satisfied  bool   // whether the version meets the constraint
	Message    string // human-readable result
}

// Checker performs environment checks against project requirements.
type Checker struct {
	requirements map[string]string
	localCfg     *localconfig.Config
}

// New creates a Checker with the given requirements and local config.
func New(requirements map[string]string, localCfg *localconfig.Config) *Checker {
	return &Checker{
		requirements: requirements,
		localCfg:     localCfg,
	}
}

// Check runs all environment checks and returns results.
func (c *Checker) Check() []CheckResult {
	results := make([]CheckResult, 0, len(c.requirements))

	for tool, constraintStr := range c.requirements {
		r := c.checkOne(tool, constraintStr)
		results = append(results, r)
	}

	return results
}

// CheckAllPassed returns true if all checks passed.
func CheckAllPassed(results []CheckResult) bool {
	for _, r := range results {
		if !r.Satisfied {
			return false
		}
	}
	return true
}

// checkOne performs a check for a single tool.
func (c *Checker) checkOne(tool, constraintStr string) CheckResult {
	r := CheckResult{
		Tool:     tool,
		Required: constraintStr,
	}

	// Parse constraint.
	constraint, err := ParseConstraint(constraintStr)
	if err != nil {
		r.Message = fmt.Sprintf("invalid constraint %q: %v", constraintStr, err)
		return r
	}

	// Resolve tool path.
	toolPath := c.resolvePath(tool)
	if toolPath == "" {
		r.Found = false
		r.Message = fmt.Sprintf("❌ %s not found (need %s)", tool, constraintStr)
		return r
	}

	r.Found = true
	r.ActualPath = toolPath

	// Get version.
	versionOutput := localconfig.DetectVersion(tool, toolPath)
	r.ActualVer = versionOutput

	versionStr := ExtractVersion(versionOutput)
	if versionStr == "" {
		r.Message = fmt.Sprintf("⚠️  %s found at %s but version unknown (need %s)", tool, toolPath, constraintStr)
		return r
	}

	r.ParsedVer = versionStr

	// Parse and compare.
	actualVer, err := ParseVersion(versionStr)
	if err != nil {
		r.Message = fmt.Sprintf("⚠️  %s version %q could not be parsed: %v", tool, versionStr, err)
		return r
	}

	r.Satisfied = constraint.Satisfied(actualVer)
	if r.Satisfied {
		r.Message = fmt.Sprintf("✅ %s %s %s (%s)", tool, versionStr, constraintStr, toolPath)
	} else {
		r.Message = fmt.Sprintf("❌ %s %s found, need %s (%s)", tool, versionStr, constraintStr, toolPath)
	}

	return r
}

// resolvePath finds the path for a tool, checking localconfig first, then PATH.
func (c *Checker) resolvePath(tool string) string {
	// Check localconfig first.
	if c.localCfg != nil {
		if p := c.localCfg.GetPath(tool); p != tool {
			// Verify it actually exists.
			if _, err := exec.LookPath(p); err == nil {
				return p
			}
			// Configured path doesn't exist, fall through to PATH search.
		}
	}

	// Fallback: search PATH.
	aliases := toolAliases(tool)
	for _, alias := range aliases {
		if p := localconfig.DetectPath(alias); p != "" {
			return p
		}
	}

	return ""
}

// toolAliases returns alternative names for a tool.
// e.g., "java" → ["java"], "maven" → ["mvn", "mvnw"]
func toolAliases(tool string) []string {
	switch strings.ToLower(tool) {
	case "maven", "mvn":
		return []string{"mvn", "mvnw"}
	case "gradle":
		return []string{"gradle", "gradlew"}
	case "python":
		return []string{"python3", "python"}
	case "pip":
		return []string{"pip3", "pip"}
	case "node":
		return []string{"node"}
	case "npm":
		return []string{"npm"}
	case "pnpm":
		return []string{"pnpm"}
	case "yarn":
		return []string{"yarn"}
	case "java":
		return []string{"java"}
	case "mysql":
		return []string{"mysql"}
	case "redis", "redis-cli":
		return []string{"redis-cli", "redis"}
	case "docker":
		return []string{"docker"}
	case "git":
		return []string{"git"}
	default:
		return []string{tool}
	}
}

// FormatResults returns a formatted table of check results.
func FormatResults(results []CheckResult) string {
	if len(results) == 0 {
		return "No requirements specified."
	}

	var b strings.Builder
	b.WriteString("Environment Check\n")
	b.WriteString("=================\n\n")

	allPassed := true
	for _, r := range results {
		b.WriteString("  ")
		b.WriteString(r.Message)
		b.WriteString("\n")
		if !r.Satisfied {
			allPassed = false
		}
	}

	b.WriteString("\n")
	if allPassed {
		b.WriteString("✅ All requirements satisfied.")
	} else {
		b.WriteString("❌ Some requirements not met. Install missing tools and try again.")
	}

	return b.String()
}

// GetWindowsExtensions returns common executable extensions on Windows.
func GetWindowsExtensions() []string {
	if runtime.GOOS != "windows" {
		return nil
	}
	return []string{"", ".exe", ".cmd", ".bat", ".ps1"}
}
