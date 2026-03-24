package orchestrator

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertgumeny/doug-plan/internal/approval"
)

const greenfieldVision = `---
project_mode: "greenfield"
language: "typescript"
runtime: "node"
---

# Vision

## Project Name

Test App

## Problem Statement

A test application for manifest sync.
`

const noFrontmatterVision = `# Vision

## Project Name

Legacy App

## Problem Statement

Some existing-project content here.
`

const missingFieldsVision = `---
project_mode: "greenfield"
---

# Vision

## Project Name

Incomplete App

## Problem Statement

Missing required scaffold fields.
`

// makeManifestTestRoot creates a temp project root with plan dir and minimal config.
func makeManifestTestRoot(t *testing.T) (root string, planDir string) {
	t.Helper()
	root = t.TempDir()
	planDir = filepath.Join(root, ".doug", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("mkdirAll planDir: %v", err)
	}
	cfg := "command:\n  - echo\napproval_mode: auto\n"
	if err := os.WriteFile(filepath.Join(planDir, "doug-plan.yaml"), []byte(cfg), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	return root, planDir
}

// makeDiscoveryAgentMock returns a mock invokeAgent that writes visionContent
// to VISION.md and records SUCCESS in ACTIVE_STEP.md.
func makeDiscoveryAgentMock(visionContent string) func(string, []string) error {
	return func(projectRoot string, _ []string) error {
		planDir := filepath.Join(projectRoot, ".doug", "plan")
		if err := os.WriteFile(filepath.Join(planDir, "VISION.md"), []byte(visionContent), 0o644); err != nil {
			return err
		}
		stepPath := filepath.Join(planDir, "ACTIVE_STEP.md")
		data, err := os.ReadFile(stepPath)
		if err != nil {
			return err
		}
		updated := strings.ReplaceAll(string(data), `outcome: ""`, `outcome: "SUCCESS"`)
		return os.WriteFile(stepPath, []byte(updated), 0o644)
	}
}

// --- applyReentry manifest cleanup tests ---

func TestApplyReentry_FreshClearsManifest(t *testing.T) {
	root, planDir := makeManifestTestRoot(t)
	manifestPath := filepath.Join(planDir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte("schema_version: 1\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	opts := Options{ProjectRoot: root, Out: io.Discard, Fresh: true}
	if err := applyReentry(opts, planDir); err != nil {
		t.Fatalf("applyReentry: %v", err)
	}
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Error("expected manifest.yaml to be removed after --fresh")
	}
}

func TestApplyReentry_FreshNoManifest_NoError(t *testing.T) {
	root, planDir := makeManifestTestRoot(t)
	opts := Options{ProjectRoot: root, Out: io.Discard, Fresh: true}
	if err := applyReentry(opts, planDir); err != nil {
		t.Fatalf("applyReentry with no manifest: %v", err)
	}
}

func TestApplyReentry_RerunDiscoveryClearsManifest(t *testing.T) {
	root, planDir := makeManifestTestRoot(t)
	if err := os.WriteFile(filepath.Join(planDir, "VISION.md"), []byte("# Vision\n\nContent.\n"), 0o644); err != nil {
		t.Fatalf("write VISION.md: %v", err)
	}
	manifestPath := filepath.Join(planDir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte("schema_version: 1\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	opts := Options{ProjectRoot: root, Out: io.Discard, RerunStage: "Discovery"}
	if err := applyReentry(opts, planDir); err != nil {
		t.Fatalf("applyReentry: %v", err)
	}
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Error("expected manifest.yaml to be removed when rerunning Discovery")
	}
}

func TestApplyReentry_RerunDiscoveryNoManifest_NoError(t *testing.T) {
	root, planDir := makeManifestTestRoot(t)
	if err := os.WriteFile(filepath.Join(planDir, "VISION.md"), []byte("# Vision\n\nContent.\n"), 0o644); err != nil {
		t.Fatalf("write VISION.md: %v", err)
	}
	opts := Options{ProjectRoot: root, Out: io.Discard, RerunStage: "Discovery"}
	if err := applyReentry(opts, planDir); err != nil {
		t.Fatalf("applyReentry with no manifest: %v", err)
	}
}

func TestApplyReentry_RerunRoadmappingPreservesManifest(t *testing.T) {
	root, planDir := makeManifestTestRoot(t)
	if err := os.WriteFile(filepath.Join(planDir, "VISION.md"), []byte("# Vision\n\nContent.\n"), 0o644); err != nil {
		t.Fatalf("write VISION.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "ROADMAP.md"), []byte("# Roadmap\n"), 0o644); err != nil {
		t.Fatalf("write ROADMAP.md: %v", err)
	}
	manifestPath := filepath.Join(planDir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte("schema_version: 1\n"), 0o644); err != nil {
		t.Fatalf("write manifest: %v", err)
	}

	opts := Options{ProjectRoot: root, Out: io.Discard, RerunStage: "Roadmapping"}
	if err := applyReentry(opts, planDir); err != nil {
		t.Fatalf("applyReentry: %v", err)
	}
	if _, err := os.Stat(manifestPath); err != nil {
		t.Errorf("manifest.yaml should be preserved when rerunning Roadmapping: %v", err)
	}
}

// --- post-Discovery manifest sync tests ---

func TestRun_ManifestSyncAfterDiscovery_Greenfield(t *testing.T) {
	root, planDir := makeManifestTestRoot(t)

	oldInvoke := invokeAgent
	oldTerminal := terminalGateFunc
	invokeAgent = makeDiscoveryAgentMock(greenfieldVision)
	terminalGateFunc = func(_ approval.Mode, _ string, _ io.Writer, _ io.Reader) error { return nil }
	defer func() {
		invokeAgent = oldInvoke
		terminalGateFunc = oldTerminal
	}()

	opts := Options{
		ProjectRoot:  root,
		Out:          io.Discard,
		In:           strings.NewReader(""),
		ApprovalMode: "auto",
	}
	if err := Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	manifestPath := filepath.Join(planDir, "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("manifest.yaml not created after greenfield Discovery: %v", err)
	}
	if !strings.Contains(string(data), "mode: greenfield") {
		t.Errorf("manifest.yaml missing greenfield mode\nGot:\n%s", string(data))
	}
	if !strings.Contains(string(data), "language: typescript") {
		t.Errorf("manifest.yaml missing language field\nGot:\n%s", string(data))
	}
}

func TestRun_ManifestSyncAfterDiscovery_NonGreenfield(t *testing.T) {
	root, planDir := makeManifestTestRoot(t)

	oldInvoke := invokeAgent
	oldTerminal := terminalGateFunc
	invokeAgent = makeDiscoveryAgentMock(noFrontmatterVision)
	terminalGateFunc = func(_ approval.Mode, _ string, _ io.Writer, _ io.Reader) error { return nil }
	defer func() {
		invokeAgent = oldInvoke
		terminalGateFunc = oldTerminal
	}()

	opts := Options{
		ProjectRoot:  root,
		Out:          io.Discard,
		In:           strings.NewReader(""),
		ApprovalMode: "auto",
	}
	if err := Run(opts); err != nil {
		t.Fatalf("Run: %v", err)
	}

	manifestPath := filepath.Join(planDir, "manifest.yaml")
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Error("manifest.yaml should not be created for non-greenfield project")
	}
}

// TestRun_RerunDiscovery_RemovesAndRegeneratesManifest verifies the full rerun
// path: a pre-existing manifest from a previous Discovery run is removed by
// applyReentry, and then regenerated after the new Discovery agent completes.
func TestRun_RerunDiscovery_RemovesAndRegeneratesManifest(t *testing.T) {
	root, planDir := makeManifestTestRoot(t)

	// Pre-populate VISION.md and manifest.yaml as if a previous Discovery run
	// had already completed and produced a manifest for a different project.
	if err := os.WriteFile(filepath.Join(planDir, "VISION.md"), []byte("# Vision\n\nOld content.\n"), 0o644); err != nil {
		t.Fatalf("write old VISION.md: %v", err)
	}
	oldManifest := "schema_version: 1\nproject:\n  name: Old Project\n  mode: greenfield\n"
	manifestPath := filepath.Join(planDir, "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte(oldManifest), 0o644); err != nil {
		t.Fatalf("write old manifest: %v", err)
	}

	oldInvoke := invokeAgent
	oldTerminal := terminalGateFunc
	invokeAgent = makeDiscoveryAgentMock(greenfieldVision)
	terminalGateFunc = func(_ approval.Mode, _ string, _ io.Writer, _ io.Reader) error { return nil }
	defer func() {
		invokeAgent = oldInvoke
		terminalGateFunc = oldTerminal
	}()

	opts := Options{
		ProjectRoot:  root,
		Out:          io.Discard,
		In:           strings.NewReader(""),
		ApprovalMode: "auto",
		RerunStage:   "Discovery",
	}
	if err := Run(opts); err != nil {
		t.Fatalf("Run with RerunStage=Discovery: %v", err)
	}

	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("manifest.yaml not present after rerun: %v", err)
	}
	content := string(data)
	if strings.Contains(content, "Old Project") {
		t.Error("manifest still contains old project name after rerun")
	}
	if !strings.Contains(content, "mode: greenfield") {
		t.Errorf("regenerated manifest missing greenfield mode\nGot:\n%s", content)
	}
	if !strings.Contains(content, "language: typescript") {
		t.Errorf("regenerated manifest missing language field\nGot:\n%s", content)
	}
}

func TestRun_ManifestSyncAfterDiscovery_MissingRequiredFields(t *testing.T) {
	root, _ := makeManifestTestRoot(t)

	oldInvoke := invokeAgent
	oldTerminal := terminalGateFunc
	invokeAgent = makeDiscoveryAgentMock(missingFieldsVision)
	terminalGateFunc = func(_ approval.Mode, _ string, _ io.Writer, _ io.Reader) error { return nil }
	defer func() {
		invokeAgent = oldInvoke
		terminalGateFunc = oldTerminal
	}()

	opts := Options{
		ProjectRoot:  root,
		Out:          io.Discard,
		In:           strings.NewReader(""),
		ApprovalMode: "cli",
	}
	err := Run(opts)
	if err == nil {
		t.Fatal("expected error for VISION.md with missing required fields, got nil")
	}
	if !strings.Contains(err.Error(), "language") || !strings.Contains(err.Error(), "runtime") {
		t.Errorf("error should mention missing fields; got: %v", err)
	}
}

// TestRun_ManifestSyncAfterDiscovery_MalformedFrontmatter_Auto verifies that
// syntactically malformed VISION.md frontmatter in auto mode produces a
// human-readable error and writes no partial manifest.
func TestRun_ManifestSyncAfterDiscovery_MalformedFrontmatter_Auto(t *testing.T) {
	root, planDir := makeManifestTestRoot(t)

	malformedVision := "---\n: invalid: yaml: [\n---\n# Vision\n\n## Project Name\n\nBad App\n"

	oldInvoke := invokeAgent
	oldTerminal := terminalGateFunc
	invokeAgent = makeDiscoveryAgentMock(malformedVision)
	terminalGateFunc = func(_ approval.Mode, _ string, _ io.Writer, _ io.Reader) error { return nil }
	defer func() {
		invokeAgent = oldInvoke
		terminalGateFunc = oldTerminal
	}()

	opts := Options{
		ProjectRoot:  root,
		Out:          io.Discard,
		In:           strings.NewReader(""),
		ApprovalMode: "auto",
	}
	err := Run(opts)
	if err == nil {
		t.Fatal("expected error for malformed VISION.md frontmatter in auto mode, got nil")
	}
	// Verify the error is human-readable (contains context about VISION.md).
	if !strings.Contains(err.Error(), "VISION.md") {
		t.Errorf("error should reference VISION.md; got: %v", err)
	}
	// No partial manifest should be written.
	manifestPath := filepath.Join(planDir, "manifest.yaml")
	if _, statErr := os.Stat(manifestPath); !os.IsNotExist(statErr) {
		t.Error("partial manifest must not be written when frontmatter is malformed")
	}
}
