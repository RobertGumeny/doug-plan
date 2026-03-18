package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/robertgumeny/doug-plan/internal/state"
)

const activeStepFile = "ACTIVE_STEP.md"

// WriteStep writes a fresh ACTIVE_STEP.md briefing to <projectRoot>/.doug/
// before the agent is invoked for the given stage.
func WriteStep(projectRoot string, stage state.Stage) error {
	dougDir := filepath.Join(projectRoot, ".doug")
	if err := os.MkdirAll(dougDir, 0o755); err != nil {
		return fmt.Errorf("creating .doug dir: %w", err)
	}

	content := fmt.Sprintf(`# Active Step

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

	dest := filepath.Join(dougDir, activeStepFile)
	if err := os.WriteFile(dest, []byte(content), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", activeStepFile, err)
	}
	return nil
}

// ArchiveStep moves <projectRoot>/.doug/ACTIVE_STEP.md into
// <projectRoot>/.doug/plans/logs/ with a unique name so archived files do not
// overwrite each other.
func ArchiveStep(projectRoot string, stage state.Stage) error {
	src := filepath.Join(projectRoot, ".doug", activeStepFile)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // nothing to archive
	}

	logsDir := filepath.Join(projectRoot, ".doug", "plans", "logs")
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
