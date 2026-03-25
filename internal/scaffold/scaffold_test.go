package scaffold_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertgumeny/doug-plan/internal/scaffold"
	"gopkg.in/yaml.v3"
)

type scaffoldConfig struct {
	Agent      string   `yaml:"agent"`
	SkillPaths []string `yaml:"skill_paths"`
}

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
		".doug/plan/ACTIVE_STEP.md",
		".doug/plan/doug-plan.yaml",
		".doug/plan/epics",
		".doug/plan/logs",
		"AGENTS.md",
		"CLAUDE.md",
		".claude/settings.json",
		".claude/skills/research/SKILL.md",
		".claude/skills/discovery/SKILL.md",
		".claude/skills/roadmapping/SKILL.md",
		".claude/skills/definition/SKILL.md",
		".claude/skills/handoff/SKILL.md",
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

func TestRun_CodexAgent(t *testing.T) {
	dir := t.TempDir()
	var out strings.Builder

	err := scaffold.Run(scaffold.Options{
		ProjectRoot: dir,
		Agents:      []string{"codex"},
		Out:         &out,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".codex", "skills", "research", "SKILL.md")); os.IsNotExist(err) {
		t.Error("expected .codex skill template to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".codex", "skills", "discovery", "SKILL.md")); os.IsNotExist(err) {
		t.Error("expected .codex discovery skill template to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".codex", "skills", "roadmapping", "SKILL.md")); os.IsNotExist(err) {
		t.Error("expected .codex roadmapping skill template to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".codex", "skills", "definition", "SKILL.md")); os.IsNotExist(err) {
		t.Error("expected .codex definition skill template to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".codex", "skills", "handoff", "SKILL.md")); os.IsNotExist(err) {
		t.Error("expected .codex handoff skill template to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "skills")); !os.IsNotExist(err) {
		t.Error("expected .claude/skills NOT to be created for codex-only init")
	}
}

func TestRun_GeminiAgent(t *testing.T) {
	dir := t.TempDir()
	var out strings.Builder

	err := scaffold.Run(scaffold.Options{
		ProjectRoot: dir,
		Agents:      []string{"gemini"},
		Out:         &out,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".gemini", "skills", "research", "SKILL.md")); os.IsNotExist(err) {
		t.Error("expected .gemini skill template to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gemini", "skills", "discovery", "SKILL.md")); os.IsNotExist(err) {
		t.Error("expected .gemini discovery skill template to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gemini", "skills", "roadmapping", "SKILL.md")); os.IsNotExist(err) {
		t.Error("expected .gemini roadmapping skill template to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gemini", "skills", "definition", "SKILL.md")); os.IsNotExist(err) {
		t.Error("expected .gemini definition skill template to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".gemini", "skills", "handoff", "SKILL.md")); os.IsNotExist(err) {
		t.Error("expected .gemini handoff skill template to be created")
	}
	if _, err := os.Stat(filepath.Join(dir, ".claude", "skills")); !os.IsNotExist(err) {
		t.Error("expected .claude/skills NOT to be created for gemini-only init")
	}
}

func TestRun_MultipleAgents(t *testing.T) {
	dir := t.TempDir()
	var out strings.Builder

	err := scaffold.Run(scaffold.Options{
		ProjectRoot: dir,
		Agents:      []string{"claude", "codex", "gemini"},
		Out:         &out,
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	for _, rel := range []string{
		".claude/skills/research/SKILL.md",
		".codex/skills/research/SKILL.md",
		".gemini/skills/research/SKILL.md",
		".claude/skills/discovery/SKILL.md",
		".codex/skills/discovery/SKILL.md",
		".gemini/skills/discovery/SKILL.md",
		".claude/skills/roadmapping/SKILL.md",
		".codex/skills/roadmapping/SKILL.md",
		".gemini/skills/roadmapping/SKILL.md",
		".claude/skills/definition/SKILL.md",
		".codex/skills/definition/SKILL.md",
		".gemini/skills/definition/SKILL.md",
		".claude/skills/handoff/SKILL.md",
		".codex/skills/handoff/SKILL.md",
		".gemini/skills/handoff/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); os.IsNotExist(err) {
			t.Errorf("expected file not created: %s", rel)
		}
	}

	config, err := os.ReadFile(filepath.Join(dir, ".doug", "plan", "doug-plan.yaml"))
	if err != nil {
		t.Fatal(err)
	}

	var cfg scaffoldConfig
	if err := yaml.Unmarshal(config, &cfg); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	for _, path := range []string{".claude/skills", ".codex/skills", ".gemini/skills"} {
		found := false
		for _, got := range cfg.SkillPaths {
			if got == path {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected skill path %q in config, got %v", path, cfg.SkillPaths)
		}
	}
	if cfg.Agent != "claude" {
		t.Errorf("primary agent = %q, want %q", cfg.Agent, "claude")
	}
}

func TestRun_SkipsExistingFiles(t *testing.T) {
	dir := t.TempDir()

	agentsPath := filepath.Join(dir, "AGENTS.md")
	customContent := "# My Custom AGENTS\n"
	if err := os.WriteFile(agentsPath, []byte(customContent), 0o644); err != nil {
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

func TestRun_NoAgentsDefaultsToClaude(t *testing.T) {
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

	for _, rel := range []string{
		".doug/plan/ACTIVE_STEP.md",
		".doug/plan/doug-plan.yaml",
		"AGENTS.md",
		"CLAUDE.md",
		".claude/skills/research/SKILL.md",
		".claude/skills/discovery/SKILL.md",
		".claude/skills/roadmapping/SKILL.md",
		".claude/skills/definition/SKILL.md",
		".claude/skills/handoff/SKILL.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); os.IsNotExist(err) {
			t.Errorf("expected file not created: %s", rel)
		}
	}
}

func TestRun_FailsWhenAlreadyInitialized(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".doug", "plan"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".doug", "plan", "doug-plan.yaml"), []byte("agent: claude\napproval_mode: auto\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := scaffold.Run(scaffold.Options{
		ProjectRoot: dir,
		Agents:      []string{"claude"},
		Out:         new(strings.Builder),
	})
	if err == nil {
		t.Fatal("expected init to fail when .doug/plan/doug-plan.yaml already exists")
	}
}

func TestRun_AllowsExistingDougDirWithoutPlanConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".doug"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	err := scaffold.Run(scaffold.Options{
		ProjectRoot: dir,
		Agents:      []string{"claude"},
		Out:         new(strings.Builder),
	})
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".doug", "plan", "doug-plan.yaml")); os.IsNotExist(err) {
		t.Fatal("expected .doug/plan/doug-plan.yaml to be created")
	}
}

func TestRun_CreatesProjectYAML(t *testing.T) {
	dir := t.TempDir()
	if err := scaffold.Run(scaffold.Options{ProjectRoot: dir, Agents: []string{"claude"}, Out: new(strings.Builder)}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".doug", "project.yaml"))
	if err != nil {
		t.Fatalf("project.yaml not created: %v", err)
	}
	if !strings.Contains(string(data), "project_id:") {
		t.Errorf("project.yaml missing project_id field, got: %q", string(data))
	}
}

func TestRun_AgentsMDContainsDougProjectID(t *testing.T) {
	dir := t.TempDir()
	if err := scaffold.Run(scaffold.Options{ProjectRoot: dir, Agents: []string{"claude"}, Out: new(strings.Builder)}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md not created: %v", err)
	}
	if !strings.Contains(string(agents), "DOUG_PROJECT_ID:") {
		t.Errorf("AGENTS.md missing DOUG_PROJECT_ID, got:\n%s", string(agents))
	}
}

func TestRun_ProjectIDIdempotent(t *testing.T) {
	dir := t.TempDir()
	if err := scaffold.Run(scaffold.Options{ProjectRoot: dir, Agents: []string{"claude"}, Out: new(strings.Builder)}); err != nil {
		t.Fatalf("first Run returned error: %v", err)
	}

	// Remove plan config so a second init is allowed
	if err := os.Remove(filepath.Join(dir, ".doug", "plan", "doug-plan.yaml")); err != nil {
		t.Fatalf("removing plan config: %v", err)
	}
	// Remove AGENTS.md so it's re-created
	if err := os.Remove(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Fatalf("removing AGENTS.md: %v", err)
	}

	if err := scaffold.Run(scaffold.Options{ProjectRoot: dir, Agents: []string{"claude"}, Out: new(strings.Builder)}); err != nil {
		t.Fatalf("second Run returned error: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ".doug", "project.yaml"))
	if err != nil {
		t.Fatalf("project.yaml not found: %v", err)
	}
	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md not found: %v", err)
	}

	// Extract project_id from project.yaml
	var idLine string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "project_id:") {
			idLine = strings.TrimSpace(strings.TrimPrefix(line, "project_id:"))
			break
		}
	}
	if idLine == "" {
		t.Fatalf("could not parse project_id from project.yaml: %q", string(data))
	}

	if !strings.Contains(string(agents), "DOUG_PROJECT_ID: "+idLine) {
		t.Errorf("AGENTS.md does not contain DOUG_PROJECT_ID: %s\ngot:\n%s", idLine, string(agents))
	}
}

func TestRun_ReadsExistingProjectYAMLID(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".doug"), 0o755); err != nil {
		t.Fatal(err)
	}
	existingID := "myproject-abc123"
	if err := os.WriteFile(filepath.Join(dir, ".doug", "project.yaml"), []byte("project_id: "+existingID+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.Run(scaffold.Options{ProjectRoot: dir, Agents: []string{"claude"}, Out: new(strings.Builder)}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md not found: %v", err)
	}
	if !strings.Contains(string(agents), "DOUG_PROJECT_ID: "+existingID) {
		t.Errorf("expected AGENTS.md to contain DOUG_PROJECT_ID: %s, got:\n%s", existingID, string(agents))
	}
}

func TestRun_MergesProjectIDIntoExistingManagedBlock(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".doug"), 0o755); err != nil {
		t.Fatal(err)
	}
	existingID := "myproject-abc123"
	if err := os.WriteFile(filepath.Join(dir, ".doug", "project.yaml"), []byte("project_id: "+existingID+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// AGENTS.md exists with a managed block but no DOUG_PROJECT_ID
	existingAgents := "# My Project\n\n<!-- DOUG-PLAN-SPECIFIC-INSTRUCTIONS:START -->\nSome content\n<!-- DOUG-PLAN-SPECIFIC-INSTRUCTIONS:END -->\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(existingAgents), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := scaffold.Run(scaffold.Options{ProjectRoot: dir, Agents: []string{"claude"}, Out: new(strings.Builder)}); err != nil {
		t.Fatalf("Run returned error: %v", err)
	}

	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md not found: %v", err)
	}
	if !strings.Contains(string(agents), "DOUG_PROJECT_ID: "+existingID) {
		t.Errorf("expected AGENTS.md to contain DOUG_PROJECT_ID: %s after merge, got:\n%s", existingID, string(agents))
	}
	if !strings.Contains(string(agents), "# My Project") {
		t.Errorf("existing content was lost:\n%s", string(agents))
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
