package scaffold

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Options holds the configuration for project scaffolding.
type Options struct {
	ProjectRoot string
	Agents      []string
	Out         io.Writer
}

// Run creates the expected directories and files for a new doug-plan project.
// Files that already exist are skipped without error.
func Run(opts Options) error {
	var created, skipped []string

	mkdir := func(rel string) error {
		path := filepath.Join(opts.ProjectRoot, rel)
		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("creating directory %s: %w", rel, err)
		}
		return nil
	}

	writeFile := func(rel, content string) error {
		path := filepath.Join(opts.ProjectRoot, rel)
		if _, err := os.Stat(path); err == nil {
			skipped = append(skipped, rel)
			return nil
		}
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return fmt.Errorf("creating parent directory for %s: %w", rel, err)
		}
		tmp := path + ".tmp"
		if err := os.WriteFile(tmp, []byte(content), 0644); err != nil {
			os.Remove(tmp)
			return fmt.Errorf("writing %s: %w", rel, err)
		}
		if err := os.Rename(tmp, path); err != nil {
			os.Remove(tmp)
			return fmt.Errorf("finalizing %s: %w", rel, err)
		}
		created = append(created, rel)
		return nil
	}

	// .doug/plans/ directory + ACTIVE_STEP.md stub
	if err := mkdir(".doug/plans"); err != nil {
		return err
	}
	if err := writeFile(".doug/plans/ACTIVE_STEP.md", activeStepStub); err != nil {
		return err
	}

	// doug-plan.yaml config
	agentStr := strings.Join(opts.Agents, ", ")
	if err := writeFile("doug-plan.yaml", buildConfig(agentStr)); err != nil {
		return err
	}

	// AGENTS.md
	if err := writeFile("AGENTS.md", agentsMDTemplate); err != nil {
		return err
	}

	// Agent-specific skill directories
	for _, agent := range opts.Agents {
		switch agent {
		case "claude":
			if err := mkdir(".claude/skills"); err != nil {
				return err
			}
			if err := writeFile(".claude/skills/.gitkeep", ""); err != nil {
				return err
			}
		}
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

func buildConfig(agentStr string) string {
	return fmt.Sprintf(`agent: %s
approval_mode: full
skill_paths:
  - .claude/skills
`, agentStr)
}

const activeStepStub = `# Active Step

No step is currently active.
`

const agentsMDTemplate = `# AGENTS.md

This file contains instructions for AI agents working in this repository.
Add project-specific rules, context, and conventions here.
`
