package orchestrator_test

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/robertgumeny/doug-plan/internal/agent"
	"github.com/robertgumeny/doug-plan/internal/orchestrator"
	"github.com/robertgumeny/doug-plan/internal/state"
)

// fakeAgentExe is the path to the compiled fake-agent binary, built once in
// TestMain and reused across all tests in this package.
var fakeAgentExe string

// TestMain compiles the testdata/fakeagent helper before running any tests.
func TestMain(m *testing.M) {
	exe, err := buildFakeAgent()
	if err != nil {
		// Non-fatal: end-to-end tests will skip themselves when fakeAgentExe == "".
		fmt.Fprintf(os.Stderr, "TestMain: build fake agent (skipping e2e tests): %v\n", err)
	} else {
		fakeAgentExe = exe
		defer func() {
			_ = os.Remove(fakeAgentExe)
		}()
	}

	os.Exit(m.Run())
}

// buildFakeAgent compiles the testdata/fakeagent package and returns the path
// to the resulting binary.
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

// TestFakeAgent_DirectInvoke manually runs the fakeagent subprocess and
// verifies ParseResult can read the updated ACTIVE_STEP.md.
func TestFakeAgent_DirectInvoke(t *testing.T) {
	if fakeAgentExe == "" {
		t.Skip("fake agent binary not built")
	}

	root := t.TempDir()
	planDir := filepath.Join(root, ".doug", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a Discovery ACTIVE_STEP.md directly.
	if err := agent.WriteStep(root, state.StageDiscovery); err != nil {
		t.Fatalf("WriteStep: %v", err)
	}

	// Invoke the fakeagent directly.
	if err := agent.Invoke(root, []string{fakeAgentExe}); err != nil {
		t.Fatalf("Invoke: %v", err)
	}

	// Read the file and log it.
	data, err := os.ReadFile(filepath.Join(planDir, "ACTIVE_STEP.md"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	t.Logf("ACTIVE_STEP.md after invoke: %q", data)

	// Call ParseResult.
	outcome, err := agent.ParseResult(root)
	t.Logf("ParseResult: outcome=%q err=%v", outcome, err)
	if err != nil {
		t.Fatalf("ParseResult failed: %v", err)
	}
}

// TestEndToEnd_DiscoveryToRoadmapping runs the orchestrator through the
// Discovery and Roadmapping stages using a fake agent subprocess, then
// validates that ROADMAP.md conforms to the hybrid Markdown + YAML format.
func TestEndToEnd_DiscoveryToRoadmapping(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode")
	}
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

	// --- Discovery stage ---
	if err := orchestrator.Run(opts); err != nil {
		t.Fatalf("Discovery stage failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(planDir, "VISION.md")); os.IsNotExist(err) {
		t.Fatal("VISION.md not created after Discovery stage")
	}

	// --- Roadmapping stage ---
	if err := orchestrator.Run(opts); err != nil {
		t.Fatalf("Roadmapping stage failed: %v", err)
	}
	roadmapPath := filepath.Join(planDir, "ROADMAP.md")
	if _, err := os.Stat(roadmapPath); os.IsNotExist(err) {
		t.Fatal("ROADMAP.md not created after Roadmapping stage")
	}

	// --- Format validation ---
	roadmapData, err := os.ReadFile(roadmapPath)
	if err != nil {
		t.Fatalf("read ROADMAP.md: %v", err)
	}
	if err := validateRoadmapFormat(string(roadmapData)); err != nil {
		t.Fatalf("ROADMAP.md format validation failed: %v", err)
	}
}

// validateRoadmapFormat checks that content conforms to the hybrid
// Markdown + YAML frontmatter format specified by the Roadmapping skill:
//
//   - Top-level YAML frontmatter with project, generated, and source fields.
//   - A "# Roadmap" heading.
//   - At least three epic sections, each with an embedded YAML block
//     containing id, name, sequence, and status fields.
//   - status is always "planned".
//   - sequence values are consecutive integers starting at 1.
func validateRoadmapFormat(content string) error {
	// Top-level frontmatter.
	if !strings.HasPrefix(content, "---\n") {
		return fmt.Errorf("missing opening frontmatter delimiter")
	}
	closingIdx := strings.Index(content[4:], "\n---\n")
	if closingIdx < 0 {
		return fmt.Errorf("missing closing frontmatter delimiter")
	}
	frontmatter := content[4 : 4+closingIdx]
	for _, field := range []string{"project:", "generated:", "source:"} {
		if !strings.Contains(frontmatter, field) {
			return fmt.Errorf("frontmatter missing %q field", field)
		}
	}
	if !strings.Contains(frontmatter, "source: VISION.md") {
		return fmt.Errorf("frontmatter source must be VISION.md")
	}

	// Document heading.
	if !strings.Contains(content, "# Roadmap") {
		return fmt.Errorf("missing '# Roadmap' heading")
	}

	// Epic sections: ## EPIC-N: Title followed by an embedded YAML block.
	epicHeading := regexp.MustCompile(`(?m)^## EPIC-\d+:`)
	epicCount := len(epicHeading.FindAllString(content, -1))
	if epicCount < 3 {
		return fmt.Errorf("expected at least 3 epics, found %d", epicCount)
	}

	// Each epic must have a parseable YAML block with the required fields.
	epicBlock := regexp.MustCompile(`(?s)## (EPIC-\d+):.*?\n---\n(.*?)\n---`)
	blocks := epicBlock.FindAllStringSubmatch(content, -1)
	if len(blocks) != epicCount {
		return fmt.Errorf("epic heading/YAML-block count mismatch: %d headings, %d blocks", epicCount, len(blocks))
	}

	for i, block := range blocks {
		yaml := block[2]
		expectedSeq := strconv.Itoa(i + 1)
		for _, field := range []string{"id:", "name:", "sequence:", "status:"} {
			if !strings.Contains(yaml, field) {
				return fmt.Errorf("epic %d YAML block missing %q field", i+1, field)
			}
		}
		if !strings.Contains(yaml, "status: planned") {
			return fmt.Errorf("epic %d status must be 'planned'", i+1)
		}
		if !strings.Contains(yaml, "sequence: "+expectedSeq) {
			return fmt.Errorf("epic %d sequence must be %s", i+1, expectedSeq)
		}
	}
	return nil
}

// TestEndToEnd_ScopingReentry verifies that when VISION.md and ROADMAP.md are
// already present, the orchestrator drops directly into Scoping — skipping
// Discovery and Roadmapping — and re-enters Scoping for each subsequent epic
// via the RETRY loop until all epics are scoped.
func TestEndToEnd_ScopingReentry(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode")
	}
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

	// Run Discovery and Roadmapping first so that VISION.md and ROADMAP.md
	// exist before Scoping begins. This simulates a real prior pipeline run.
	for _, stage := range []string{"Discovery", "Roadmapping"} {
		out.Reset()
		if err := orchestrator.Run(opts); err != nil {
			t.Fatalf("%s stage failed: %v", stage, err)
		}
	}

	// All subsequent runs must re-enter at Scoping — never at Discovery or
	// Roadmapping — because VISION.md and ROADMAP.md are already present.
	// fakeRoadmap defines 4 epics; the first 3 runs return RETRY and the 4th
	// writes global SCOPED.md and returns SUCCESS.
	const numEpics = 4
	for i := range numEpics {
		out.Reset()
		if err := orchestrator.Run(opts); err != nil {
			t.Fatalf("run %d: orchestrator.Run failed: %v", i+1, err)
		}
		output := out.String()

		// Every run must enter at Scoping — never at Discovery or Roadmapping.
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

	// After all epics are scoped the global SCOPED.md must exist.
	if _, err := os.Stat(filepath.Join(planDir, "SCOPED.md")); os.IsNotExist(err) {
		t.Fatal("SCOPED.md not created after all epics scoped")
	}

	// Every per-epic SCOPED.md must exist.
	for _, id := range []string{"EPIC-1", "EPIC-2", "EPIC-3", "EPIC-4"} {
		path := filepath.Join(planDir, "epics", id, "SCOPED.md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("per-epic SCOPED.md missing for %s", id)
		}
	}
}

// TestEndToEnd_FullPipeline runs the orchestrator through all pipeline stages —
// Discovery, Roadmapping, Scoping (with re-entry), and PRD/Handoff (with
// re-entry) — using the fake agent subprocess, and verifies all output
// artifacts are present at the end.
func TestEndToEnd_FullPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping end-to-end test in short mode")
	}
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

	// Discovery
	if err := orchestrator.Run(opts); err != nil {
		t.Fatalf("Discovery stage failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(planDir, "VISION.md")); os.IsNotExist(err) {
		t.Fatal("VISION.md not created after Discovery")
	}

	// Roadmapping
	if err := orchestrator.Run(opts); err != nil {
		t.Fatalf("Roadmapping stage failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(planDir, "ROADMAP.md")); os.IsNotExist(err) {
		t.Fatal("ROADMAP.md not created after Roadmapping")
	}

	// Scoping — fakeRoadmap has 4 epics; each run scopes one, returning RETRY
	// until the last epic, at which point SCOPED.md is written and SUCCESS returned.
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

	// Handoff (PRD stage) — one run per epic; last run writes global PRD.md.
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

// TestValidateRoadmapFormat_Valid verifies the validator accepts a correctly
// formatted ROADMAP.md.
func TestValidateRoadmapFormat_Valid(t *testing.T) {
	content := `---
project: "Test Project"
generated: "2026-03-18"
source: VISION.md
---

# Roadmap

## EPIC-1: First Epic

---
id: EPIC-1
name: "First Epic"
sequence: 1
status: planned
---

Description of first epic.

## EPIC-2: Second Epic

---
id: EPIC-2
name: "Second Epic"
sequence: 2
status: planned
---

Description of second epic.

## EPIC-3: Third Epic

---
id: EPIC-3
name: "Third Epic"
sequence: 3
status: planned
---

Description of third epic.
`
	if err := validateRoadmapFormat(content); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

// TestValidateRoadmapFormat_Invalid verifies the validator rejects malformed content.
func TestValidateRoadmapFormat_Invalid(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{
			name:    "missing frontmatter",
			content: "# Roadmap\n\n## EPIC-1: Foo\n",
		},
		{
			name: "missing project field",
			content: `---
generated: "2026-01-01"
source: VISION.md
---

# Roadmap

## EPIC-1: A

---
id: EPIC-1
name: "A"
sequence: 1
status: planned
---

## EPIC-2: B

---
id: EPIC-2
name: "B"
sequence: 2
status: planned
---

## EPIC-3: C

---
id: EPIC-3
name: "C"
sequence: 3
status: planned
---
`,
		},
		{
			name: "source not VISION.md",
			content: `---
project: "X"
generated: "2026-01-01"
source: OTHER.md
---

# Roadmap

## EPIC-1: A

---
id: EPIC-1
name: "A"
sequence: 1
status: planned
---

## EPIC-2: B

---
id: EPIC-2
name: "B"
sequence: 2
status: planned
---

## EPIC-3: C

---
id: EPIC-3
name: "C"
sequence: 3
status: planned
---
`,
		},
		{
			name: "fewer than 3 epics",
			content: `---
project: "X"
generated: "2026-01-01"
source: VISION.md
---

# Roadmap

## EPIC-1: Only One

---
id: EPIC-1
name: "Only One"
sequence: 1
status: planned
---

One epic is not enough.
`,
		},
		{
			name: "status not planned",
			content: `---
project: "X"
generated: "2026-01-01"
source: VISION.md
---

# Roadmap

## EPIC-1: A

---
id: EPIC-1
name: "A"
sequence: 1
status: done
---

## EPIC-2: B

---
id: EPIC-2
name: "B"
sequence: 2
status: planned
---

## EPIC-3: C

---
id: EPIC-3
name: "C"
sequence: 3
status: planned
---
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateRoadmapFormat(tc.content); err == nil {
				t.Errorf("expected validation error for %q, got nil", tc.name)
			}
		})
	}
}
