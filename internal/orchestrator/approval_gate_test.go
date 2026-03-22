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

func TestRunApprovalGate_HardModeUsesBrowserGate(t *testing.T) {
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

	err := runApprovalGate(Options{}, &config.Config{ApprovalMode: "hard"}, state.StagePRD, plansDir)
	if err != nil {
		t.Fatalf("runApprovalGate: %v", err)
	}
	if !browserCalled {
		t.Fatal("expected browser gate to be called")
	}
	if terminalCalled {
		t.Fatal("terminal gate should not be called in hard mode")
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
	err := runApprovalGate(
		Options{ApprovalMode: "soft", Out: &out},
		&config.Config{ApprovalMode: "hard"},
		state.StageRoadmapping,
		t.TempDir(),
	)
	if err != nil {
		t.Fatalf("runApprovalGate: %v", err)
	}
	if gotMode != approval.ModeSoft {
		t.Fatalf("mode = %q, want %q", gotMode, approval.ModeSoft)
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

	err := runApprovalGate(Options{}, &config.Config{}, state.StageDiscovery, t.TempDir())
	if err != nil {
		t.Fatalf("runApprovalGate: %v", err)
	}
}
