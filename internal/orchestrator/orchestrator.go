package orchestrator

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/robertgumeny/doug-plan/internal/state"
)

// Options holds the runtime configuration for an orchestrator run.
type Options struct {
	ProjectRoot string
	Out         io.Writer
}

// Run infers the current pipeline position from artifacts on disk and
// reports the entry point. No state file is read or written.
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

	fmt.Fprintf(opts.Out, "Pipeline entry point: %s\n", stage)
	return nil
}
