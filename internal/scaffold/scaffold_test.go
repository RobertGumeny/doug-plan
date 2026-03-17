package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertgumeny/doug-plan/internal/scaffold"
)

func TestRun_CreatesExpectedFiles(t *testing.T) {
	dir := t.TempDir()
	var out strings.Builder

	err := scaffold.Run(scaffold.Options{
		ProjectRoot: dir,
		Agents:      []string{"claude"},
		Out:         &out,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	expected := []string{
		".doug/plans/ACTIVE_STEP.md",
		"doug-plan.yaml",
		"AGENTS.md",
		".claude/skills/.gitkeep",
	}
	for _, rel := range expected {
		path := filepath.Join(dir, rel)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("expected file not created: %s", rel)
		}
	}

	output := out.String()
	if !strings.Contains(output, "doug-plan init complete") {
		t.Errorf("output missing summary line, got: %q", output)
	}
}

func TestRun_SkipsExistingFiles(t *testing.T) {
	dir := t.TempDir()

	// Pre-create AGENTS.md with custom content
	agentsPath := filepath.Join(dir, "AGENTS.md")
	customContent := "# My Custom AGENTS\n"
	if err := os.WriteFile(agentsPath, []byte(customContent), 0644); err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	err := scaffold.Run(scaffold.Options{
		ProjectRoot: dir,
		Agents:      []string{"claude"},
		Out:         &out,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	got, _ := os.ReadFile(agentsPath)
	if string(got) != customContent {
		t.Errorf("existing AGENTS.md was overwritten; got %q", string(got))
	}

	if !strings.Contains(out.String(), "Skipped") {
		t.Errorf("expected 'Skipped' in output, got: %q", out.String())
	}
}

func TestRun_NoAgents(t *testing.T) {
	dir := t.TempDir()
	var out strings.Builder

	err := scaffold.Run(scaffold.Options{
		ProjectRoot: dir,
		Agents:      nil,
		Out:         &out,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	// Core files should still be created
	for _, rel := range []string{".doug/plans/ACTIVE_STEP.md", "doug-plan.yaml", "AGENTS.md"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); os.IsNotExist(err) {
			t.Errorf("expected file not created: %s", rel)
		}
	}
}

func TestParseAgents(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"single agent", "claude", []string{"claude"}},
		{"multiple agents", "claude, codex", []string{"claude", "codex"}},
		{"empty string", "", nil},
		{"whitespace only", "  ", nil},
		{"spaces around commas", " claude , codex ", []string{"claude", "codex"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := scaffold.ParseAgents(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("ParseAgents(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("ParseAgents(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}
