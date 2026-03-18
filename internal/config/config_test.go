package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad(t *testing.T) {
	root := t.TempDir()
	yaml := "agent: claude\napproval_mode: full\nskill_paths:\n  - .claude/skills\n"
	if err := os.WriteFile(filepath.Join(root, "doug-plan.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	cfg, err := Load(root)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Agent != "claude" {
		t.Errorf("Agent = %q, want %q", cfg.Agent, "claude")
	}
	if cfg.ApprovalMode != "full" {
		t.Errorf("ApprovalMode = %q, want %q", cfg.ApprovalMode, "full")
	}
}

func TestLoad_MissingFile(t *testing.T) {
	root := t.TempDir()
	_, err := Load(root)
	if err == nil {
		t.Fatal("expected error for missing doug-plan.yaml, got nil")
	}
}

func TestAgentCommand_DefaultClaude(t *testing.T) {
	cfg := &Config{Agent: "claude"}
	args, err := cfg.AgentCommand()
	if err != nil {
		t.Fatalf("AgentCommand: %v", err)
	}
	if len(args) == 0 || args[0] != "claude" {
		t.Errorf("args[0] = %q, want %q", args[0], "claude")
	}
}

func TestAgentCommand_DefaultCodex(t *testing.T) {
	cfg := &Config{Agent: "codex"}
	args, err := cfg.AgentCommand()
	if err != nil {
		t.Fatalf("AgentCommand: %v", err)
	}
	if len(args) == 0 || args[0] != "codex" {
		t.Errorf("args[0] = %q, want %q", args[0], "codex")
	}
}

func TestAgentCommand_DefaultGemini(t *testing.T) {
	cfg := &Config{Agent: "gemini"}
	args, err := cfg.AgentCommand()
	if err != nil {
		t.Fatalf("AgentCommand: %v", err)
	}
	if len(args) == 0 || args[0] != "gemini" {
		t.Errorf("args[0] = %q, want %q", args[0], "gemini")
	}
}

func TestAgentCommand_CustomCommand(t *testing.T) {
	cfg := &Config{
		Agent:   "claude",
		Command: []string{"my-agent", "--flag", "value"},
	}
	args, err := cfg.AgentCommand()
	if err != nil {
		t.Fatalf("AgentCommand: %v", err)
	}
	if len(args) != 3 || args[0] != "my-agent" {
		t.Errorf("args = %v, want [my-agent --flag value]", args)
	}
}

func TestAgentCommand_UnknownAgent(t *testing.T) {
	cfg := &Config{Agent: "unknown-bot"}
	_, err := cfg.AgentCommand()
	if err == nil {
		t.Fatal("expected error for unknown agent, got nil")
	}
}
