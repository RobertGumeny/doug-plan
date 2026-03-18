package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the orchestrator configuration from doug-plan.yaml.
type Config struct {
	Agent        string   `yaml:"agent"`
	Command      []string `yaml:"command,omitempty"`
	ApprovalMode string   `yaml:"approval_mode"`
	SkillPaths   []string `yaml:"skill_paths"`
}

// Load reads and parses doug-plan.yaml from the project root.
func Load(projectRoot string) (*Config, error) {
	path := filepath.Join(projectRoot, "doug-plan.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading doug-plan.yaml: %w", err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing doug-plan.yaml: %w", err)
	}
	return &cfg, nil
}

// AgentCommand returns the command args to invoke the agent.
// If cfg.Command is set, it is used directly. Otherwise a default
// is derived from cfg.Agent.
func (c *Config) AgentCommand() ([]string, error) {
	if len(c.Command) > 0 {
		return c.Command, nil
	}
	switch c.Agent {
	case "claude":
		return []string{"claude", "--print", "Please complete the step described in .doug/ACTIVE_STEP.md"}, nil
	case "codex":
		return []string{"codex", "Please complete the step described in .doug/ACTIVE_STEP.md"}, nil
	case "gemini":
		return []string{"gemini", "Please complete the step described in .doug/ACTIVE_STEP.md"}, nil
	default:
		return nil, fmt.Errorf("unknown agent %q: set command in doug-plan.yaml", c.Agent)
	}
}
