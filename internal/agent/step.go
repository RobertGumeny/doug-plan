package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/robertgumeny/doug-plan/internal/layout"
	"github.com/robertgumeny/doug-plan/internal/state"
	"github.com/robertgumeny/doug-plan/internal/templates"
)

const activeStepFile = layout.ActiveStepFile

// WriteStep writes a fresh ACTIVE_STEP.md briefing to <projectRoot>/.doug/plan/
// before the agent is invoked for the given stage.
//
// Stage-specific templates in internal/templates/steps/<Stage>.md are used when
// available; otherwise a generic template is written.
func WriteStep(projectRoot string, stage state.Stage) error {
	planDir := layout.PlanDir(projectRoot)
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		return fmt.Errorf("creating plan dir: %w", err)
	}

	content := stepContent(stage)

	dest := layout.ActiveStepPath(projectRoot)
	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", activeStepFile, err)
	}
	return nil
}

// stepContent returns the ACTIVE_STEP.md content for the given stage.
// It loads a stage-specific template from the embedded steps FS when one
// exists, and falls back to a generic template for stages without one.
func stepContent(stage state.Stage) string {
	name := "steps/" + stage.String() + ".md"
	if data, err := templates.Steps.ReadFile(name); err == nil {
		return string(data)
	}
	return fmt.Sprintf(`# Active Step

**Stage**: %s

## Briefing

Execute the %s step of the pipeline and produce its output artifact.

---

## Agent Result

---
outcome: ""
---

## Output
`, stage, stage)
}

// ArchiveStep moves <projectRoot>/.doug/plan/ACTIVE_STEP.md into
// <projectRoot>/.doug/plan/logs/ with a unique name so archived files do not
// overwrite each other.
func ArchiveStep(projectRoot string, stage state.Stage) error {
	src := layout.ActiveStepPath(projectRoot)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // nothing to archive
	}

	logsDir := layout.LogsDir(projectRoot)
	if err := os.MkdirAll(logsDir, 0o755); err != nil {
		return fmt.Errorf("creating logs dir: %w", err)
	}

	name := fmt.Sprintf("%s_%d.md", stage, time.Now().UnixNano())
	dest := filepath.Join(logsDir, name)

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading %s: %w", activeStepFile, err)
	}
	if err := os.WriteFile(dest, data, 0o644); err != nil {
		return fmt.Errorf("writing archive %s: %w", name, err)
	}
	if err := os.Remove(src); err != nil {
		return fmt.Errorf("removing %s: %w", activeStepFile, err)
	}
	return nil
}
