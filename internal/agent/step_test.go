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
		"VISION.md",     // artifact path
		"/discovery",    // skill instruction
		"outcome: \"\"", // result stub
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
		"ROADMAP.md",    // artifact path
		"/roadmapping",  // skill instruction
		"VISION.md",     // prerequisite
		"outcome: \"\"", // result stub
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("Roadmapping ACTIVE_STEP.md missing %q; got:\n%s", want, content)
		}
	}
}

func TestWriteStep_Definition_TemplateContent(t *testing.T) {
	root := t.TempDir()

	if err := WriteStep(root, state.StageDefinition); err != nil {
		t.Fatalf("WriteStep: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".doug", "plan", activeStepFile))
	if err != nil {
		t.Fatalf("reading ACTIVE_STEP.md: %v", err)
	}

	content := string(data)
	checks := []string{
		"DEFINITION.md", // artifact path
		"/definition",   // skill instruction
		"ROADMAP.md",    // prerequisite
		"VISION.md",     // prerequisite
		"outcome: \"\"", // result stub
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("Definition ACTIVE_STEP.md missing %q; got:\n%s", want, content)
		}
	}
}

func TestWriteStep_PRD_TemplateContent(t *testing.T) {
	root := t.TempDir()

	if err := WriteStep(root, state.StagePRD); err != nil {
		t.Fatalf("WriteStep: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(root, ".doug", "plan", activeStepFile))
	if err != nil {
		t.Fatalf("reading ACTIVE_STEP.md: %v", err)
	}

	content := string(data)
	checks := []string{
		"PRD.md",        // artifact path
		"/handoff",      // skill instruction
		"DEFINITION.md", // prerequisite
		"outcome: \"\"", // result stub
	}
	for _, want := range checks {
		if !strings.Contains(content, want) {
			t.Errorf("PRD ACTIVE_STEP.md missing %q; got:\n%s", want, content)
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

func TestMaterializeArtifact_CreatesShell(t *testing.T) {
	cases := []struct {
		stage        state.Stage
		artifactFile string
		wantHeading  string
	}{
		{state.StageDiscovery, "VISION.md", "# Vision"},
		{state.StageRoadmapping, "ROADMAP.md", "# Roadmap"},
		{state.StageDefinition, "DEFINITION.md", "# Definition"},
	}

	for _, tc := range cases {
		t.Run(tc.artifactFile, func(t *testing.T) {
			root := t.TempDir()

			if err := MaterializeArtifact(root, tc.stage); err != nil {
				t.Fatalf("MaterializeArtifact: %v", err)
			}

			dest := filepath.Join(root, ".doug", "plan", tc.artifactFile)
			data, err := os.ReadFile(dest)
			if err != nil {
				t.Fatalf("reading %s: %v", tc.artifactFile, err)
			}
			if !strings.Contains(string(data), tc.wantHeading) {
				t.Errorf("%s missing heading %q; got:\n%s", tc.artifactFile, tc.wantHeading, data)
			}
		})
	}
}

func TestMaterializeArtifact_NoOverwriteExisting(t *testing.T) {
	root := t.TempDir()

	planDir := filepath.Join(root, ".doug", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	existing := "agent-written content"
	dest := filepath.Join(planDir, "VISION.md")
	if err := os.WriteFile(dest, []byte(existing), 0o644); err != nil {
		t.Fatalf("writing existing file: %v", err)
	}

	if err := MaterializeArtifact(root, state.StageDiscovery); err != nil {
		t.Fatalf("MaterializeArtifact: %v", err)
	}

	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading VISION.md: %v", err)
	}
	if string(data) != existing {
		t.Errorf("expected existing content to be preserved; got:\n%s", data)
	}
}

func TestMaterializeArtifact_NoopForStagesWithoutTemplate(t *testing.T) {
	root := t.TempDir()

	// PRD and Tasks stages have no artifact templates; should be a clean no-op.
	for _, stage := range []state.Stage{state.StagePRD, state.StageTasks} {
		if err := MaterializeArtifact(root, stage); err != nil {
			t.Fatalf("MaterializeArtifact(%s): %v", stage, err)
		}
		// Artifact file must not have been created.
		artifactFile := state.ArtifactFile(stage)
		dest := filepath.Join(root, ".doug", "plan", artifactFile)
		if _, err := os.Stat(dest); !os.IsNotExist(err) {
			t.Errorf("expected %s not to exist after no-op MaterializeArtifact", artifactFile)
		}
	}
}
