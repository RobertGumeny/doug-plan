package scaffold

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/robertgumeny/doug-plan/internal/layout"
	"github.com/robertgumeny/doug-plan/internal/templates"
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

	fmt.Fprintf(opts.Out, "doug-plan init complete\n\n")
	if len(created) > 0 {
		fmt.Fprintf(opts.Out, "Created:\n")
		for _, f := range created {
			fmt.Fprintf(opts.Out, "  %s\n", f)
		}
	}
	if len(skipped) > 0 {
		fmt.Fprintf(opts.Out, "Skipped (already exist):\n")
		for _, f := range skipped {
			fmt.Fprintf(opts.Out, "  %s\n", f)
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

		var dst string
		switch {
		case rel == "AGENTS.md":
			dst = filepath.Join(projectRoot, "AGENTS.md")
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
