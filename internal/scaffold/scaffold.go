package scaffold

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/robertgumeny/doug-plan/internal/layout"
	"github.com/robertgumeny/doug-plan/internal/templates"
	"gopkg.in/yaml.v3"
)

// Options holds the configuration for project scaffolding.
type Options struct {
	ProjectRoot string
	Agents      []string
	Out         io.Writer
}

// Run creates the expected directories and files for a new doug-plan project.
// If .doug/plan/doug-plan.yaml already exists, the project is treated as already initialized.
func Run(opts Options) error {
	var created, skipped []string

	if _, err := os.Stat(layout.ConfigPath(opts.ProjectRoot)); err == nil {
		return fmt.Errorf("%s already exists — project appears to already be initialized", layout.ConfigPath(opts.ProjectRoot))
	}

	agents := normalizeAgents(opts.Agents)

	for _, dir := range []string{
		layout.DougDir(opts.ProjectRoot),
		layout.PlanDir(opts.ProjectRoot),
		layout.LogsDir(opts.ProjectRoot),
		layout.EpicsDir(opts.ProjectRoot),
	} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory %s: %w", dir, err)
		}
	}

	if err := writeManagedFile(layout.ActiveStepPath(opts.ProjectRoot), []byte(activeStepStub), opts.ProjectRoot, &created, &skipped); err != nil {
		return err
	}
	if err := writeManagedFile(layout.ConfigPath(opts.ProjectRoot), []byte(buildConfig(agents)), opts.ProjectRoot, &created, &skipped); err != nil {
		return err
	}
	if err := copyInitTemplates(opts.ProjectRoot, agents, &created, &skipped); err != nil {
		return err
	}

	writef(opts.Out, "doug-plan init complete\n\n")
	if len(created) > 0 {
		writef(opts.Out, "Created:\n")
		for _, f := range created {
			writef(opts.Out, "  %s\n", f)
		}
	}
	if len(skipped) > 0 {
		writef(opts.Out, "Skipped (already exist):\n")
		for _, f := range skipped {
			writef(opts.Out, "  %s\n", f)
		}
	}

	return nil
}

// ParseAgents splits a comma-separated agents string into a trimmed slice.
// Empty input returns an empty slice.
func ParseAgents(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func normalizeAgents(agents []string) []string {
	if len(agents) == 0 {
		return []string{"claude"}
	}

	out := make([]string, 0, len(agents))
	seen := make(map[string]struct{}, len(agents))
	for _, agent := range agents {
		agent = strings.ToLower(strings.TrimSpace(agent))
		if agent == "" {
			continue
		}
		if _, ok := seen[agent]; ok {
			continue
		}
		seen[agent] = struct{}{}
		out = append(out, agent)
	}

	if len(out) == 0 {
		return []string{"claude"}
	}
	return out
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func buildConfig(agents []string) string {
	var skillPaths strings.Builder
	for _, agent := range agents {
		switch agent {
		case "claude":
			skillPaths.WriteString("  - .claude/skills\n")
		case "codex":
			skillPaths.WriteString("  - .codex/skills\n")
		case "gemini":
			skillPaths.WriteString("  - .gemini/skills\n")
		}
	}

	primaryAgent := "claude"
	if len(agents) > 0 && agents[0] != "" {
		primaryAgent = agents[0]
	}

	return fmt.Sprintf("agent: %s\napproval_mode: auto\nskill_paths:\n%s", primaryAgent, skillPaths.String())
}

func copyInitTemplates(projectRoot string, agents []string, created, skipped *[]string) error {
	agentSelected := make(map[string]bool, len(agents))
	for _, agent := range agents {
		agentSelected[agent] = true
	}

	projectID, err := readOrCreateProjectYAML(projectRoot)
	if err != nil {
		return fmt.Errorf("reading or creating project.yaml: %w", err)
	}

	return fs.WalkDir(templates.Init, "init", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		rel := strings.TrimPrefix(path, "init/")
		data, err := templates.Init.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading template %s: %w", rel, err)
		}

		if strings.HasPrefix(rel, "skills/") {
			skillRel := strings.TrimPrefix(rel, "skills/")
			for _, dst := range selectedSkillDestinations(projectRoot, agentSelected, skillRel) {
				if err := writeManagedFile(dst, data, projectRoot, created, skipped); err != nil {
					return err
				}
			}
			return nil
		}

		if rel == "AGENTS.md" {
			return copyOrMergePlanAgents(projectRoot, projectID, data, created, skipped)
		}

		var dst string
		switch {
		case rel == "CLAUDE.md":
			dst = filepath.Join(projectRoot, "CLAUDE.md")
		case strings.HasPrefix(rel, ".claude/"):
			if !agentSelected["claude"] {
				return nil
			}
			dst = filepath.Join(projectRoot, rel)
		case strings.HasPrefix(rel, ".codex/"):
			if !agentSelected["codex"] {
				return nil
			}
			dst = filepath.Join(projectRoot, rel)
		case strings.HasPrefix(rel, ".gemini/"):
			if !agentSelected["gemini"] {
				return nil
			}
			dst = filepath.Join(projectRoot, rel)
		default:
			return nil
		}

		return writeManagedFile(dst, data, projectRoot, created, skipped)
	})
}

func selectedSkillDestinations(projectRoot string, agentSelected map[string]bool, skillRel string) []string {
	var destinations []string
	if agentSelected["claude"] {
		destinations = append(destinations, filepath.Join(projectRoot, ".claude", "skills", skillRel))
	}
	if agentSelected["codex"] {
		destinations = append(destinations, filepath.Join(projectRoot, ".codex", "skills", skillRel))
	}
	if agentSelected["gemini"] {
		destinations = append(destinations, filepath.Join(projectRoot, ".gemini", "skills", skillRel))
	}
	return destinations
}

type projectYAML struct {
	ProjectID string `yaml:"project_id"`
}

// readOrCreateProjectYAML returns the project_id from .doug/project.yaml,
// creating the file with a generated ID if it does not already exist.
func readOrCreateProjectYAML(projectRoot string) (string, error) {
	path := layout.ProjectYAMLPath(projectRoot)

	if data, err := os.ReadFile(path); err == nil {
		var p projectYAML
		if err := yaml.Unmarshal(data, &p); err == nil && p.ProjectID != "" {
			return p.ProjectID, nil
		}
	}

	id, err := generateProjectID(projectRoot)
	if err != nil {
		return "", fmt.Errorf("generating project ID: %w", err)
	}

	p := projectYAML{ProjectID: id}
	data, err := yaml.Marshal(p)
	if err != nil {
		return "", fmt.Errorf("marshaling project.yaml: %w", err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("writing project.yaml: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return "", fmt.Errorf("finalizing project.yaml: %w", err)
	}

	return id, nil
}

var slugRe = regexp.MustCompile(`[^a-z0-9]+`)

func generateProjectID(projectRoot string) (string, error) {
	base := strings.ToLower(filepath.Base(projectRoot))
	base = slugRe.ReplaceAllString(base, "-")
	base = strings.Trim(base, "-")
	if base == "" {
		base = "project"
	}

	var b [3]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%x", base, b), nil
}

const (
	managedBlockStart = "<!-- DOUG-PLAN-SPECIFIC-INSTRUCTIONS:START -->"
	managedBlockEnd   = "<!-- DOUG-PLAN-SPECIFIC-INSTRUCTIONS:END -->"
	projectIDPrefix   = "DOUG_PROJECT_ID:"
)

// copyOrMergePlanAgents writes the AGENTS.md template with the project ID
// injected when the file does not exist. When the file already exists it
// merges the DOUG_PROJECT_ID line into the managed block idempotently.
// Files that exist but have no managed block are added to skipped unchanged.
func copyOrMergePlanAgents(projectRoot, projectID string, templateData []byte, created, skipped *[]string) error {
	dst := filepath.Join(projectRoot, "AGENTS.md")
	content := strings.ReplaceAll(string(templateData), "{{DOUG_PROJECT_ID}}", projectID)

	if _, err := os.Stat(dst); os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("creating parent directory for AGENTS.md: %w", err)
		}
		tmp := dst + ".tmp"
		if err := os.WriteFile(tmp, []byte(content), 0o644); err != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("writing AGENTS.md: %w", err)
		}
		if err := os.Rename(tmp, dst); err != nil {
			_ = os.Remove(tmp)
			return fmt.Errorf("finalizing AGENTS.md: %w", err)
		}
		rel, relErr := filepath.Rel(projectRoot, dst)
		if relErr != nil {
			rel = dst
		}
		*created = append(*created, rel)
		return nil
	}

	existing, err := os.ReadFile(dst)
	if err != nil {
		return fmt.Errorf("reading AGENTS.md: %w", err)
	}

	merged, ok := mergeProjectID(string(existing), projectID)
	if !ok {
		rel, relErr := filepath.Rel(projectRoot, dst)
		if relErr != nil {
			rel = dst
		}
		*skipped = append(*skipped, rel)
		return nil
	}

	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, []byte(merged), 0o644); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("writing AGENTS.md: %w", err)
	}
	if err := os.Rename(tmp, dst); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("finalizing AGENTS.md: %w", err)
	}
	return nil
}

// mergeProjectID finds the managed block in content and sets DOUG_PROJECT_ID
// to projectID within it. Returns the updated content and true on success, or
// the original content and false when no managed block is found.
func mergeProjectID(content, projectID string) (string, bool) {
	startIdx := strings.Index(content, managedBlockStart)
	if startIdx < 0 {
		return content, false
	}
	endIdx := strings.Index(content, managedBlockEnd)
	if endIdx < 0 || endIdx < startIdx {
		return content, false
	}

	blockInner := content[startIdx+len(managedBlockStart) : endIdx]
	newLine := projectIDPrefix + " " + projectID

	lines := strings.Split(blockInner, "\n")
	found := false
	for i, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), projectIDPrefix) {
			lines[i] = newLine
			found = true
			break
		}
	}

	var newInner string
	if found {
		newInner = strings.Join(lines, "\n")
	} else {
		// Insert as first line after the start marker
		newInner = "\n" + newLine + blockInner
	}

	result := content[:startIdx] + managedBlockStart + newInner + managedBlockEnd + content[endIdx+len(managedBlockEnd):]
	return result, true
}

func writeManagedFile(path string, data []byte, projectRoot string, created, skipped *[]string) error {
	if _, err := os.Stat(path); err == nil {
		rel, relErr := filepath.Rel(projectRoot, path)
		if relErr != nil {
			rel = path
		}
		*skipped = append(*skipped, rel)
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating parent directory for %s: %w", path, err)
	}

	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("writing %s: %w", path, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("finalizing %s: %w", path, err)
	}

	rel, relErr := filepath.Rel(projectRoot, path)
	if relErr != nil {
		rel = path
	}
	*created = append(*created, rel)
	return nil
}

const activeStepStub = `# Active Step

No step is currently active.
`
