package orchestrator

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/robertgumeny/doug-plan/internal/agent"
	"github.com/robertgumeny/doug-plan/internal/state"
)

// Options holds the runtime configuration for an orchestrator run.
type Options struct {
	ProjectRoot string
	Out         io.Writer
}

// Run infers the current pipeline position from artifacts on disk,
// writes an ACTIVE_STEP.md briefing, and archives it after the step completes.
func Run(opts Options) error {
	plansDir := filepath.Join(opts.ProjectRoot, ".doug", "plans")

	stage, err := state.InferStage(plansDir)
	if err != nil {
		return fmt.Errorf("inferring pipeline stage: %w", err)
	}

	if stage == state.StageComplete {
		fmt.Fprintf(opts.Out, "Pipeline complete. All artifacts are present in %s.\n", plansDir)
		return nil
	}

	if err := agent.WriteStep(opts.ProjectRoot, stage); err != nil {
		return fmt.Errorf("writing ACTIVE_STEP.md: %w", err)
	}

	fmt.Fprintf(opts.Out, "Pipeline entry point: %s\n", stage)

	if err := agent.ArchiveStep(opts.ProjectRoot, stage); err != nil {
		return fmt.Errorf("archiving ACTIVE_STEP.md: %w", err)
	}

	return nil
}
