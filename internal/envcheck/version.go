package envcheck

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Constraint represents a version constraint like ">=17" or ">=3.8.0".
type Constraint struct {
	Op    string // ">=", "<=", "==", ">", "<"
	Major int
	Minor int
	Patch int
}

// ParseConstraint parses a constraint string like ">=17", ">=3.8.0", "17".
func ParseConstraint(s string) (Constraint, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Constraint{}, fmt.Errorf("empty constraint")
	}

	// Extract operator.
	var op string
	rest := s
	for _, prefix := range []string{">=", "<=", "==", ">", "<", "="} {
		if strings.HasPrefix(rest, prefix) {
			op = prefix
			rest = strings.TrimPrefix(rest, prefix)
			break
		}
	}
	if op == "" {
		op = ">=" // default: minimum version
	}

	major, minor, patch, err := parseVersion(rest)
	if err != nil {
		return Constraint{}, fmt.Errorf("parsing version %q: %w", rest, err)
	}

	return Constraint{Op: op, Major: major, Minor: minor, Patch: patch}, nil
}

// Version represents a parsed version number.
type Version struct {
	Major int
	Minor int
	Patch int
}

// parseVersion parses "17", "3.8", "3.8.1" into major, minor, patch.
func parseVersion(s string) (major, minor, patch int, err error) {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return 0, 0, 0, fmt.Errorf("invalid version format: %s", s)
	}

	major, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, 0, fmt.Errorf("invalid major version: %s", parts[0])
	}
	if len(parts) > 1 {
		minor, err = strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid minor version: %s", parts[1])
		}
	}
	if len(parts) > 2 {
		patch, err = strconv.Atoi(parts[2])
		if err != nil {
			return 0, 0, 0, fmt.Errorf("invalid patch version: %s", parts[2])
		}
	}
	return major, minor, patch, nil
}

// ParseVersion parses a version string into a Version struct.
func ParseVersion(s string) (Version, error) {
	major, minor, patch, err := parseVersion(s)
	if err != nil {
		return Version{}, err
	}
	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// Satisfied checks if the given version satisfies this constraint.
func (c Constraint) Satisfied(v Version) bool {
	switch c.Op {
	case ">=":
		return cmpVersion(v, Version{c.Major, c.Minor, c.Patch}) >= 0
	case "<=":
		return cmpVersion(v, Version{c.Major, c.Minor, c.Patch}) <= 0
	case ">":
		return cmpVersion(v, Version{c.Major, c.Minor, c.Patch}) > 0
	case "<":
		return cmpVersion(v, Version{c.Major, c.Minor, c.Patch}) < 0
	case "==", "=":
		return cmpVersion(v, Version{c.Major, c.Minor, c.Patch}) == 0
	default:
		return false
	}
}

// cmpVersion compares two versions. Returns -1, 0, or 1.
func cmpVersion(a, b Version) int {
	if a.Major != b.Major {
		return sign(a.Major - b.Major)
	}
	if a.Minor != b.Minor {
		return sign(a.Minor - b.Minor)
	}
	if a.Patch != b.Patch {
		return sign(a.Patch - b.Patch)
	}
	return 0
}

func sign(n int) int {
	if n > 0 {
		return 1
	}
	if n < 0 {
		return -1
	}
	return 0
}

// ExtractVersion tries to extract a version number from a tool's output string.
// Handles common formats:
//   - "java version "17.0.1"" → "17.0.1"
//   - "node v22.0.0" → "22.0.0"
//   - "Maven 3.8.1" → "3.8.1"
//   - "mysql Ver 8.0.30" → "8.0.30"
//   - "go version go1.21.0" → "1.21.0"
func ExtractVersion(output string) string {
	// Try to find version patterns like X.Y.Z or X.Y or vX.Y.Z
	re := regexp.MustCompile(`(\d+\.\d+\.\d+)`)
	if m := re.FindString(output); m != "" {
		return m
	}

	// Try X.Y
	re = regexp.MustCompile(`(\d+\.\d+)`)
	if m := re.FindString(output); m != "" {
		return m
	}

	// Try bare major (e.g., "go1.21" → "1.21")
	re = regexp.MustCompile(`\D(\d+\.\d+)\b`)
	if m := re.FindString(output); m != "" {
		return strings.TrimLeft(m, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ _-")
	}

	// Try "version 17" or "v17" pattern (major only).
	re = regexp.MustCompile(`[vV]?ersion\s*[vV]?(\d+)`)
	if m := re.FindStringSubmatch(output); len(m) > 1 {
		return m[1]
	}

	re = regexp.MustCompile(`\b[vV](\d+)\b`)
	if m := re.FindStringSubmatch(output); len(m) > 1 {
		return m[1]
	}

	return ""
}

// String returns a human-readable representation of the constraint.
func (c Constraint) String() string {
	return fmt.Sprintf("%s%d.%d.%d", c.Op, c.Major, c.Minor, c.Patch)
}
