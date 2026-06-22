package envcheck

import (
	"pitstop/internal/localconfig"
	"testing"
)

func TestParseConstraint(t *testing.T) {
	tests := []struct {
		input   string
		wantOp  string
		wantMaj int
		wantMin int
		wantPat int
		wantErr bool
	}{
		{">=17", ">=", 17, 0, 0, false},
		{">=3.8", ">=", 3, 8, 0, false},
		{">=3.8.1", ">=", 3, 8, 1, false},
		{"17", ">=", 17, 0, 0, false},
		{"==1.0", "==", 1, 0, 0, false},
		{">1.0", ">", 1, 0, 0, false},
		{"<20", "<", 20, 0, 0, false},
		{"", "", 0, 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			c, err := ParseConstraint(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if c.Op != tt.wantOp {
				t.Errorf("op: got %q, want %q", c.Op, tt.wantOp)
			}
			if c.Major != tt.wantMaj || c.Minor != tt.wantMin || c.Patch != tt.wantPat {
				t.Errorf("version: got %d.%d.%d, want %d.%d.%d",
					c.Major, c.Minor, c.Patch, tt.wantMaj, tt.wantMin, tt.wantPat)
			}
		})
	}
}

func TestConstraint_Satisfied(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		{">=17", "17.0.1", true},
		{">=17", "16.0.0", false},
		{">=17", "21.0.0", true},
		{">=3.8", "3.8.1", true},
		{">=3.8", "3.9.0", true},
		{">=3.8", "3.7.0", false},
		{">=22.0.0", "22.0.0", true},
		{">=22.0.0", "21.9.9", false},
		{">1.0", "1.0.0", false},
		{">1.0", "1.0.1", true},
		{"<20", "19.9.9", true},
		{"<20", "20.0.0", false},
		{"==1.0", "1.0.0", true},
		{"==1.0", "1.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.constraint+"_"+tt.version, func(t *testing.T) {
			c, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("parse constraint: %v", err)
			}
			v, err := ParseVersion(tt.version)
			if err != nil {
				t.Fatalf("parse version: %v", err)
			}
			got := c.Satisfied(v)
			if got != tt.want {
				t.Errorf("Satisfied(%s, %s) = %v, want %v", tt.constraint, tt.version, got, tt.want)
			}
		})
	}
}

func TestExtractVersion(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{`openjdk version "17.0.1" 2021-10-19`, "17.0.1"},
		{"v22.0.0", "22.0.0"},
		{"Maven 3.8.1", "3.8.1"},
		{"mysql  Ver 8.0.30 for Linux on x86_64", "8.0.30"},
		{"go version go1.21.0 linux/amd64", "1.21.0"},
		{"node v18.17.0", "18.17.0"},
		{"pnpm 8.6.0", "8.6.0"},
		{"git version 2.39.2", "2.39.2"},
		{"Python 3.11.4", "3.11.4"},
		{"redis-cli 7.0.11", "7.0.11"},
		{"no version here", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input[:min(20, len(tt.input))], func(t *testing.T) {
			got := ExtractVersion(tt.input)
			if got != tt.want {
				t.Errorf("ExtractVersion(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCheck_Found(t *testing.T) {
	// git should be available on most dev machines.
	localCfg, _ := localconfig.AutoDetect()
	checker := New(map[string]string{"git": ">=2.0"}, localCfg)
	results := checker.Check()

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	r := results[0]
	if !r.Found {
		t.Skip("git not found on PATH, skipping")
	}
	if !r.Satisfied {
		t.Errorf("expected git >= 2.0 to be satisfied, got %s", r.ActualVer)
	}
}

func TestCheck_NotFound(t *testing.T) {
	localCfg := &localconfig.Config{Paths: map[string]string{}}
	checker := New(map[string]string{"this_tool_does_not_exist": ">=1.0"}, localCfg)
	results := checker.Check()

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Found {
		t.Error("expected tool to not be found")
	}
	if results[0].Satisfied {
		t.Error("expected Satisfied=false for missing tool")
	}
}

func TestCheckAllPassed(t *testing.T) {
	passed := []CheckResult{{Satisfied: true}, {Satisfied: true}}
	if !CheckAllPassed(passed) {
		t.Error("expected all passed")
	}

	failed := []CheckResult{{Satisfied: true}, {Satisfied: false}}
	if CheckAllPassed(failed) {
		t.Error("expected not all passed")
	}

	empty := []CheckResult{}
	if !CheckAllPassed(empty) {
		t.Error("expected empty slice to pass")
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
