package orchestrator

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/robertgumeny/doug-plan/internal/approval"
	"github.com/robertgumeny/doug-plan/internal/config"
	"github.com/robertgumeny/doug-plan/internal/state"
)

func TestRunApprovalGate_BrowserModeUsesBrowserGate(t *testing.T) {
	plansDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(plansDir, "PRD.md"), []byte("prd"), 0o644); err != nil {
		t.Fatalf("write PRD.md: %v", err)
	}
	if err := os.WriteFile(filepath.Join(plansDir, "tasks.yaml"), []byte("tasks"), 0o644); err != nil {
		t.Fatalf("write tasks.yaml: %v", err)
	}

	var gotPrimary string
	var gotSecondary string
	var gotStage string
	browserCalled := false
	terminalCalled := false

	oldBrowserGate := browserGateFunc
	oldTerminalGate := terminalGateFunc
	browserGateFunc = func(primaryPath, secondaryPath, stage string, out io.Writer) error {
		browserCalled = true
		gotPrimary = primaryPath
		gotSecondary = secondaryPath
		gotStage = stage
		return nil
	}
	terminalGateFunc = func(mode approval.Mode, stage string, out io.Writer, in io.Reader) error {
		terminalCalled = true
		return nil
	}
	defer func() {
		browserGateFunc = oldBrowserGate
		terminalGateFunc = oldTerminalGate
	}()

	browserMode, err := runApprovalGate(Options{}, &config.Config{ApprovalMode: "browser"}, state.StagePRD, plansDir)
	if err != nil {
		t.Fatalf("runApprovalGate: %v", err)
	}
	if !browserMode {
		t.Fatal("expected hardModeUsed=true for browser approval mode")
	}
	if !browserCalled {
		t.Fatal("expected browser gate to be called")
	}
	if terminalCalled {
		t.Fatal("terminal gate should not be called in browser mode")
	}
	if gotPrimary != filepath.Join(plansDir, "PRD.md") {
		t.Fatalf("primaryPath = %q", gotPrimary)
	}
	if gotSecondary != filepath.Join(plansDir, "tasks.yaml") {
		t.Fatalf("secondaryPath = %q", gotSecondary)
	}
	if gotStage != "PRD" {
		t.Fatalf("stage = %q, want %q", gotStage, "PRD")
	}
}

func TestRunApprovalGate_CLIOverrideUsesTerminalGate(t *testing.T) {
	oldBrowserGate := browserGateFunc
	oldTerminalGate := terminalGateFunc
	browserGateFunc = func(primaryPath, secondaryPath, stage string, out io.Writer) error {
		t.Fatal("browser gate should not be called")
		return nil
	}
	var gotMode approval.Mode
	var gotStage string
	terminalGateFunc = func(mode approval.Mode, stage string, out io.Writer, in io.Reader) error {
		gotMode = mode
		gotStage = stage
		return nil
	}
	defer func() {
		browserGateFunc = oldBrowserGate
		terminalGateFunc = oldTerminalGate
	}()

	var out bytes.Buffer
	hardMode, err := runApprovalGate(
		Options{ApprovalMode: "cli", Out: &out},
		&config.Config{ApprovalMode: "browser"},
		state.StageRoadmapping,
		t.TempDir(),
	)
	if err != nil {
		t.Fatalf("runApprovalGate: %v", err)
	}
	if hardMode {
		t.Fatal("expected hardModeUsed=false when CLI override is cli")
	}
	if gotMode != approval.ModeCLI {
		t.Fatalf("mode = %q, want %q", gotMode, approval.ModeCLI)
	}
	if gotStage != "Roadmapping" {
		t.Fatalf("stage = %q, want %q", gotStage, "Roadmapping")
	}
}

func TestRunApprovalGate_DefaultsToAutoTerminalGate(t *testing.T) {
	oldTerminalGate := terminalGateFunc
	terminalGateFunc = func(mode approval.Mode, stage string, out io.Writer, in io.Reader) error {
		if mode != approval.ModeAuto {
			t.Fatalf("mode = %q, want %q", mode, approval.ModeAuto)
		}
		if stage != "Discovery" {
			t.Fatalf("stage = %q, want %q", stage, "Discovery")
		}
		return nil
	}
	defer func() {
		terminalGateFunc = oldTerminalGate
	}()

	_, err := runApprovalGate(Options{}, &config.Config{}, state.StageDiscovery, t.TempDir())
	if err != nil {
		t.Fatalf("runApprovalGate: %v", err)
	}
}

// TestRunApprovalGate_DiscoveryBrowserModePassesManifestAsSecondary verifies that
// browser-mode Discovery approval passes manifest.yaml as the secondary artifact and
// returns hardModeUsed=true.
func TestRunApprovalGate_DiscoveryBrowserModePassesManifestAsSecondary(t *testing.T) {
	// Set up a minimal project root so manifest.Sync (best-effort) does not panic.
	projectRoot := t.TempDir()
	plansDir := filepath.Join(projectRoot, ".doug", "plan")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatalf("mkdir plansDir: %v", err)
	}
	// Write a minimal VISION.md artifact shell so ArtifactFile resolves.
	if err := os.WriteFile(filepath.Join(plansDir, "VISION.md"), []byte("# Vision\n"), 0o644); err != nil {
		t.Fatalf("write VISION.md: %v", err)
	}

	var gotPrimary, gotSecondary, gotStage string
	oldBrowserGate := browserGateFunc
	browserGateFunc = func(primaryPath, secondaryPath, stage string, out io.Writer) error {
		gotPrimary = primaryPath
		gotSecondary = secondaryPath
		gotStage = stage
		return nil
	}
	defer func() { browserGateFunc = oldBrowserGate }()

	hardMode, err := runApprovalGate(
		Options{ProjectRoot: projectRoot},
		&config.Config{ApprovalMode: "browser"},
		state.StageDiscovery,
		plansDir,
	)
	if err != nil {
		t.Fatalf("runApprovalGate: %v", err)
	}
	if !hardMode {
		t.Fatal("expected hardModeUsed=true for browser mode Discovery")
	}
	if gotPrimary != filepath.Join(plansDir, "VISION.md") {
		t.Fatalf("primaryPath = %q, want VISION.md", gotPrimary)
	}
	wantSecondary := filepath.Join(plansDir, "manifest.yaml")
	if gotSecondary != wantSecondary {
		t.Fatalf("secondaryPath = %q, want %q", gotSecondary, wantSecondary)
	}
	if gotStage != "Discovery" {
		t.Fatalf("stage = %q, want %q", gotStage, "Discovery")
	}
}

// TestRunApprovalGate_DiscoveryBrowserMode_MalformedFrontmatter_DoesNotCrash
// verifies that syntactically malformed VISION.md frontmatter does not cause
// runApprovalGate to return an error in browser mode. The manifest.Sync call is
// best-effort; the browser gate must still be invoked so the user can fix
// VISION.md in the split-view UI.
func TestRunApprovalGate_DiscoveryBrowserMode_MalformedFrontmatter_DoesNotCrash(t *testing.T) {
	projectRoot := t.TempDir()
	plansDir := filepath.Join(projectRoot, ".doug", "plan")
	if err := os.MkdirAll(plansDir, 0o755); err != nil {
		t.Fatalf("mkdir plansDir: %v", err)
	}
	// Write a VISION.md with syntactically malformed frontmatter.
	malformed := "---\n: invalid: yaml: [\n---\n# Vision\n"
	if err := os.WriteFile(filepath.Join(plansDir, "VISION.md"), []byte(malformed), 0o644); err != nil {
		t.Fatalf("write VISION.md: %v", err)
	}

	browserCalled := false
	oldBrowserGate := browserGateFunc
	browserGateFunc = func(_, _, _ string, _ io.Writer) error {
		browserCalled = true
		return nil
	}
	defer func() { browserGateFunc = oldBrowserGate }()

	hardMode, err := runApprovalGate(
		Options{ProjectRoot: projectRoot},
		&config.Config{ApprovalMode: "browser"},
		state.StageDiscovery,
		plansDir,
	)
	if err != nil {
		t.Fatalf("runApprovalGate should not error for malformed frontmatter in browser mode: %v", err)
	}
	if !hardMode {
		t.Fatal("expected hardModeUsed=true")
	}
	if !browserCalled {
		t.Fatal("browser gate must be called even when VISION.md frontmatter is malformed")
	}
}
