//go:build integration

package orchestrator_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertgumeny/doug-plan/internal/agent"
	"github.com/robertgumeny/doug-plan/internal/orchestrator"
	"github.com/robertgumeny/doug-plan/internal/state"
)

var fakeAgentExe string

func TestMain(m *testing.M) {
	exe, err := buildFakeAgent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "TestMain: build fake agent (skipping e2e tests): %v\n", err)
	} else {
		fakeAgentExe = exe
		defer func() {
			_ = os.Remove(fakeAgentExe)
		}()
	}

	os.Exit(m.Run())
}

func buildFakeAgent() (string, error) {
	dir, err := os.MkdirTemp("", "fakeagent-*")
	if err != nil {
		return "", fmt.Errorf("mktemp: %w", err)
	}
	exe := filepath.Join(dir, "fakeagent")
	cmd := exec.Command("go", "build", "-o", exe, "./testdata/fakeagent")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		_ = os.RemoveAll(dir)
		return "", fmt.Errorf("go build fakeagent: %w", err)
	}
	return exe, nil
}

func TestFakeAgent_DirectInvoke(t *testing.T) {
	if fakeAgentExe == "" {
		t.Skip("fake agent binary not built")
	}

	root := t.TempDir()
	planDir := filepath.Join(root, ".doug", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := agent.WriteStep(root, state.StageDiscovery); err != nil {
		t.Fatalf("WriteStep: %v", err)
	}

	if err := agent.Invoke(root, []string{fakeAgentExe}); err != nil {
		t.Fatalf("Invoke: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(planDir, "ACTIVE_STEP.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	t.Logf("ACTIVE_STEP.md after invoke: %q", data)

	outcome, err := agent.ParseResult(root)
	t.Logf("ParseResult: outcome=%q err=%v", outcome, err)
	if err != nil {
		t.Fatalf("ParseResult failed: %v", err)
	}
}

func TestEndToEnd_DiscoveryToRoadmapping(t *testing.T) {
	if fakeAgentExe == "" {
		t.Skip("fake agent binary not built")
	}

	root := t.TempDir()
	planDir := filepath.Join(root, ".doug", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("setup: mkdir planDir: %v", err)
	}

	config := fmt.Sprintf("command:\n  - %s\napproval_mode: auto\n", fakeAgentExe)
	if err := os.WriteFile(filepath.Join(planDir, "doug-plan.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("setup: write config: %v", err)
	}

	opts := orchestrator.Options{
		ProjectRoot:  root,
		Out:          io.Discard,
		In:           strings.NewReader(""),
		ApprovalMode: "auto",
	}

	if err := orchestrator.Run(opts); err != nil {
		t.Fatalf("Discovery stage failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(planDir, "VISION.md")); os.IsNotExist(err) {
		t.Fatal("VISION.md not created after Discovery stage")
	}

	if err := orchestrator.Run(opts); err != nil {
		t.Fatalf("Roadmapping stage failed: %v", err)
	}
	roadmapPath := filepath.Join(planDir, "ROADMAP.md")
	if _, err := os.Stat(roadmapPath); os.IsNotExist(err) {
		t.Fatal("ROADMAP.md not created after Roadmapping stage")
	}

	roadmapData, err := os.ReadFile(roadmapPath)
	if err != nil {
		t.Fatalf("read ROADMAP.md: %v", err)
	}
	if err := validateRoadmapFormat(string(roadmapData)); err != nil {
		t.Fatalf("ROADMAP.md format validation failed: %v", err)
	}
}

func TestEndToEnd_ScopingReentry(t *testing.T) {
	if fakeAgentExe == "" {
		t.Skip("fake agent binary not built")
	}

	root := t.TempDir()
	planDir := filepath.Join(root, ".doug", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("setup: mkdir planDir: %v", err)
	}

	config := fmt.Sprintf("command:\n  - %s\napproval_mode: auto\n", fakeAgentExe)
	if err := os.WriteFile(filepath.Join(planDir, "doug-plan.yaml"), []byte(config), 0o644); err != nil {
		t.Fatalf("setup: write config: %v", err)
	}

	var out strings.Builder
	opts := orchestrator.Options{
		ProjectRoot:  root,
		Out:          &out,
		In:           strings.NewReader(""),
		ApprovalMode: "auto",
	}

	for _, stage := range []string{"Discovery", "Roadmapping"} {
		out.Reset()
		if err := orchestrator.Run(opts); err != nil {
			t.Fatalf("%s stage failed: %v", stage, err)
		}
	}

	const numEpics = 4
	for i := range numEpics {
		out.Reset()
		if err := orchestrator.Run(opts); err != nil {
			t.Fatalf("run %d: orchestrator.Run failed: %v", i+1, err)
		}
		output := out.String()

		if !strings.Contains(output, "Pipeline entry point: Scoping") {
			t.Errorf("run %d: expected Scoping entry point, got:\n%s", i+1, output)
		}
		if strings.Contains(output, "Pipeline entry point: Discovery") {
			t.Errorf("run %d: Discovery was unexpectedly re-run:\n%s", i+1, output)
		}
		if strings.Contains(output, "Pipeline entry point: Roadmapping") {
			t.Errorf("run %d: Roadmapping was unexpectedly re-run:\n%s", i+1, output)
		}
	}

	if _, err := os.Stat(filepath.Join(planDir, "SCOPED.md")); os.IsNotExist(err) {
		t.Fatal("SCOPED.md not created after all epics scoped")
	}

	for _, id := range []string{"EPIC-1", "EPIC-2", "EPIC-3", "EPIC-4"} {
		path := filepath.Join(planDir, "epics", id, "SCOPED.md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("per-epic SCOPED.md missing for %s", id)
		}
	}
}

func TestEndToEnd_FullPipeline(t *testing.T) {
	if fakeAgentExe == "" {
		t.Skip("fake agent binary not built")
	}

	root := t.TempDir()
	planDir := filepath.Join(root, ".doug", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("setup: mkdir planDir: %v", err)
	}

	cfgContent := fmt.Sprintf("command:\n  - %s\napproval_mode: auto\n", fakeAgentExe)
	if err := os.WriteFile(filepath.Join(planDir, "doug-plan.yaml"), []byte(cfgContent), 0o644); err != nil {
		t.Fatalf("setup: write config: %v", err)
	}

	opts := orchestrator.Options{
		ProjectRoot:  root,
		Out:          io.Discard,
		In:           strings.NewReader(""),
		ApprovalMode: "auto",
	}

	if err := orchestrator.Run(opts); err != nil {
		t.Fatalf("Discovery stage failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(planDir, "VISION.md")); os.IsNotExist(err) {
		t.Fatal("VISION.md not created after Discovery")
	}

	if err := orchestrator.Run(opts); err != nil {
		t.Fatalf("Roadmapping stage failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(planDir, "ROADMAP.md")); os.IsNotExist(err) {
		t.Fatal("ROADMAP.md not created after Roadmapping")
	}

	const numEpics = 4
	for i := range numEpics {
		if err := orchestrator.Run(opts); err != nil {
			t.Fatalf("Scoping run %d failed: %v", i+1, err)
		}
	}
	if _, err := os.Stat(filepath.Join(planDir, "SCOPED.md")); os.IsNotExist(err) {
		t.Fatal("SCOPED.md not created after all Scoping runs")
	}
	for _, id := range []string{"EPIC-1", "EPIC-2", "EPIC-3", "EPIC-4"} {
		if _, err := os.Stat(filepath.Join(planDir, "epics", id, "SCOPED.md")); os.IsNotExist(err) {
			t.Errorf("per-epic SCOPED.md missing for %s", id)
		}
	}

	for i := range numEpics {
		if err := orchestrator.Run(opts); err != nil {
			t.Fatalf("Handoff run %d failed: %v", i+1, err)
		}
	}
	if _, err := os.Stat(filepath.Join(planDir, "PRD.md")); os.IsNotExist(err) {
		t.Fatal("PRD.md not created after all Handoff runs")
	}
	for _, id := range []string{"EPIC-1", "EPIC-2", "EPIC-3", "EPIC-4"} {
		epicDir := filepath.Join(planDir, "epics", id)
		if _, err := os.Stat(filepath.Join(epicDir, "PRD.md")); os.IsNotExist(err) {
			t.Errorf("per-epic PRD.md missing for %s", id)
		}
		if _, err := os.Stat(filepath.Join(epicDir, "tasks.yaml")); os.IsNotExist(err) {
			t.Errorf("per-epic tasks.yaml missing for %s", id)
		}
	}
}
