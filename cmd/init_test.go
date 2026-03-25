package cmd

import (
	"bytes"
	"strings"
	"testing"
)

func TestResolveAgents_FlagProvided_SkipsPrompt(t *testing.T) {
	result, err := resolveAgents("codex", new(bytes.Buffer), strings.NewReader(""), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "codex" {
		t.Errorf("expected [codex], got %v", result)
	}
}

func TestResolveAgents_FlagProvided_TTYStillSkipsPrompt(t *testing.T) {
	var out bytes.Buffer
	result, err := resolveAgents("codex", &out, strings.NewReader("1\n"), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "codex" {
		t.Errorf("expected [codex], got %v", result)
	}
	if out.Len() != 0 {
		t.Errorf("expected no output when flag is set, got: %s", out.String())
	}
}

func TestResolveAgents_NonTTY_DefaultsClaude(t *testing.T) {
	result, err := resolveAgents("", new(bytes.Buffer), strings.NewReader(""), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "claude" {
		t.Errorf("expected [claude], got %v", result)
	}
}

func TestResolveAgents_TTY_PromptShownAndSelectionReturned(t *testing.T) {
	var out bytes.Buffer
	result, err := resolveAgents("", &out, strings.NewReader("2\n"), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "codex" {
		t.Errorf("expected [codex], got %v", result)
	}
	if !strings.Contains(out.String(), "Select providers") {
		t.Errorf("expected prompt in output, got: %s", out.String())
	}
}

func TestResolveAgents_TTY_SelectMultipleProviders(t *testing.T) {
	var out bytes.Buffer
	// Toggle claude (1) and codex (2), then confirm with empty line.
	result, err := resolveAgents("", &out, strings.NewReader("1\n2\n\n"), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 providers, got %v", result)
	}
	if result[0] != "claude" || result[1] != "codex" {
		t.Errorf("expected [claude codex], got %v", result)
	}
}

func TestResolveAgents_TTY_EnterReturnsClaude(t *testing.T) {
	result, err := resolveAgents("", new(bytes.Buffer), strings.NewReader("\n"), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "claude" {
		t.Errorf("expected [claude] on Enter, got %v", result)
	}
}

func TestResolveAgents_TTY_SelectGemini(t *testing.T) {
	result, err := resolveAgents("", new(bytes.Buffer), strings.NewReader("3\n"), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "gemini" {
		t.Errorf("expected [gemini], got %v", result)
	}
}
