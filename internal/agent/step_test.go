package agent

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertgumeny/doug-plan/internal/state"
)

func TestWriteStep(t *testing.T) {
	root := t.TempDir()

	if err := WriteStep(root, state.StageDiscovery); err != nil {
		t.Fatalf("WriteStep: %v", err)
	}

	dest := filepath.Join(root, ".doug", "plan", activeStepFile)
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading ACTIVE_STEP.md: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "Discovery") {
		t.Errorf("ACTIVE_STEP.md missing stage name; got:\n%s", content)
	}
	if !strings.Contains(content, "## Agent Result") {
		t.Errorf("ACTIVE_STEP.md missing Agent Result section; got:\n%s", content)
	}
}

func TestWriteStep_Discovery_TemplateContent(t *testing.T) {
	root := t.TempDir()

	if err := WriteStep(root, state.StageDiscovery); err != nil {
		t.Fatalf("WriteStep: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".doug", "plan", activeStepFile))
	if err != nil {
		t.Fatalf("reading ACTIVE_STEP.md: %v", err)
	}

	content := string(data)
	checks := []string{
		"VISION.md",      // artifact path
		"/discovery",     // skill instruction
		"outcome: \"\"",  // result stub
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("Discovery ACTIVE_STEP.md missing %q; got:\n%s", want, content)
		}
	}
}

func TestWriteStep_Roadmapping_TemplateContent(t *testing.T) {
	root := t.TempDir()

	if err := WriteStep(root, state.StageRoadmapping); err != nil {
		t.Fatalf("WriteStep: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".doug", "plan", activeStepFile))
	if err != nil {
		t.Fatalf("reading ACTIVE_STEP.md: %v", err)
	}

	content := string(data)
	checks := []string{
		"ROADMAP.md",     // artifact path
		"/roadmapping",   // skill instruction
		"VISION.md",      // prerequisite
		"outcome: \"\"",  // result stub
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("Roadmapping ACTIVE_STEP.md missing %q; got:\n%s", want, content)
		}
	}
}

func TestWriteStep_OverwritesExisting(t *testing.T) {
	root := t.TempDir()

	if err := WriteStep(root, state.StageDiscovery); err != nil {
		t.Fatalf("first WriteStep: %v", err)
	}
	if err := WriteStep(root, state.StageRoadmapping); err != nil {
		t.Fatalf("second WriteStep: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".doug", "plan", activeStepFile))
	if err != nil {
		t.Fatalf("reading file: %v", err)
	}
	if !strings.Contains(string(data), "Roadmapping") {
		t.Errorf("expected Roadmapping in overwritten file; got:\n%s", data)
	}
}

func TestArchiveStep(t *testing.T) {
	root := t.TempDir()

	if err := WriteStep(root, state.StageDiscovery); err != nil {
		t.Fatalf("WriteStep: %v", err)
	}

	if err := ArchiveStep(root, state.StageDiscovery); err != nil {
		t.Fatalf("ArchiveStep: %v", err)
	}

	// ACTIVE_STEP.md should be gone
	_, err := os.Stat(filepath.Join(root, ".doug", "plan", activeStepFile))
	if !os.IsNotExist(err) {
		t.Errorf("expected ACTIVE_STEP.md to be removed after archive, got stat err: %v", err)
	}

	// archive file should exist under .doug/plan/logs/
	logsDir := filepath.Join(root, ".doug", "plan", "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		t.Fatalf("reading logs dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 archived file, got %d", len(entries))
	}
	if !strings.HasPrefix(entries[0].Name(), "Discovery_") {
		t.Errorf("unexpected archive name %q", entries[0].Name())
	}
}

func TestArchiveStep_NoOverwrite(t *testing.T) {
	root := t.TempDir()

	// Archive the same stage twice; names must differ.
	for i := 0; i < 2; i++ {
		if err := WriteStep(root, state.StagePRD); err != nil {
			t.Fatalf("WriteStep %d: %v", i, err)
		}
		if err := ArchiveStep(root, state.StagePRD); err != nil {
			t.Fatalf("ArchiveStep %d: %v", i, err)
		}
	}

	logsDir := filepath.Join(root, ".doug", "plan", "logs")
	entries, err := os.ReadDir(logsDir)
	if err != nil {
		t.Fatalf("reading logs dir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 distinct archive files, got %d", len(entries))
	}
	if entries[0].Name() == entries[1].Name() {
		t.Errorf("archive files must not have the same name: %q", entries[0].Name())
	}
}

func TestArchiveStep_Noop_WhenNoActiveStep(t *testing.T) {
	root := t.TempDir()

	// No ACTIVE_STEP.md written; archive should succeed silently.
	if err := ArchiveStep(root, state.StageDiscovery); err != nil {
		t.Fatalf("ArchiveStep with no file: %v", err)
	}
}
