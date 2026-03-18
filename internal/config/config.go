package config

import (
	"fmt"
	"os"

	"github.com/robertgumeny/doug-plan/internal/layout"
	"gopkg.in/yaml.v3"
)

// Config holds the orchestrator configuration from .doug/plan/doug-plan.yaml.
type Config struct {
	Agent        string   `yaml:"agent"`
	Command      []string `yaml:"command,omitempty"`
	ApprovalMode string   `yaml:"approval_mode"`
	SkillPaths   []string `yaml:"skill_paths"`
}

// Load reads and parses .doug/plan/doug-plan.yaml from the project root.
func Load(projectRoot string) (*Config, error) {
	path := layout.ConfigPath(projectRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", layout.ConfigFileName, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", layout.ConfigFileName, err)
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
		return []string{"claude", "--print", "Please complete the step described in .doug/plan/ACTIVE_STEP.md"}, nil
	case "codex":
		return []string{"codex", "Please complete the step described in .doug/plan/ACTIVE_STEP.md"}, nil
	case "gemini":
		return []string{"gemini", "Please complete the step described in .doug/plan/ACTIVE_STEP.md"}, nil
	default:
		return nil, fmt.Errorf("unknown agent %q: set command in .doug/plan/%s", c.Agent, layout.ConfigFileName)
	}
}
