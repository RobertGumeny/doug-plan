package orchestrator

import (
	"errors"
	"fmt"
	"io"

	"github.com/robertgumeny/doug-plan/internal/agent"
	"github.com/robertgumeny/doug-plan/internal/approval"
	"github.com/robertgumeny/doug-plan/internal/config"
	"github.com/robertgumeny/doug-plan/internal/layout"
	"github.com/robertgumeny/doug-plan/internal/state"
)

// Options holds the runtime configuration for an orchestrator run.
type Options struct {
	ProjectRoot  string
	Out          io.Writer
	In           io.Reader
	ApprovalMode string // overrides config when non-empty; must be "auto", "soft", or "hard"
	RerunStage   string // non-empty triggers re-run mode: clears this stage's artifact and all subsequent
	Fresh        bool   // true triggers start-fresh mode: clears all plan artifacts
}

// Run infers the current pipeline position from artifacts on disk,
// writes an ACTIVE_STEP.md briefing, invokes the configured agent,
// parses the result, dispatches by outcome, and archives the step file.
func Run(opts Options) error {
	plansDir := layout.PlanDir(opts.ProjectRoot)

	if err := applyReentry(opts, plansDir); err != nil {
		return err
	}

	stage, err := state.InferStage(plansDir)
	if err != nil {
		return fmt.Errorf("inferring pipeline stage: %w", err)
	}

	if stage == state.StageComplete {
		writef(opts.Out, "Pipeline complete. All artifacts are present in %s.\n", plansDir)
		return nil
	}

	if err := agent.WriteStep(opts.ProjectRoot, stage); err != nil {
		return fmt.Errorf("writing ACTIVE_STEP.md: %w", err)
	}

	writef(opts.Out, "Pipeline entry point: %s\n", stage)

	cfg, err := config.Load(opts.ProjectRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	args, err := cfg.AgentCommand()
	if err != nil {
		return fmt.Errorf("resolving agent command: %w", err)
	}

	writef(opts.Out, "Invoking agent: %s\n", args[0])
	if err := agent.Invoke(opts.ProjectRoot, args); err != nil {
		return fmt.Errorf("agent invocation: %w", err)
	}

	outcome, err := agent.ParseResult(opts.ProjectRoot)
	if err != nil {
		return fmt.Errorf("parsing agent result: %w", err)
	}

	if err := agent.ArchiveStep(opts.ProjectRoot, stage); err != nil {
		return fmt.Errorf("archiving ACTIVE_STEP.md: %w", err)
	}

	switch outcome {
	case agent.OutcomeSuccess:
		writef(opts.Out, "Step %s completed successfully.\n", stage)
		if err := runApprovalGate(opts, cfg, stage.String()); err != nil {
			if errors.Is(err, approval.ErrSkipped) {
				return nil
			}
			return err
		}
	case agent.OutcomeFailure:
		return fmt.Errorf("step %s failed: agent reported FAILURE", stage)
	case agent.OutcomeRetry:
		writef(opts.Out, "Step %s requesting retry.\n", stage)
	}

	return nil
}

// applyReentry clears plan artifacts according to the requested re-entry mode
// before the orchestrator infers the current pipeline position.
//
// Modes:
//   - Fresh (opts.Fresh == true): remove all pipeline artifacts; run begins at Discovery.
//   - Re-run (opts.RerunStage != ""): remove the named stage's artifact and every
//     subsequent stage's artifact so the run starts at that stage.
//   - Resume (neither flag set): no-op; InferStage picks up where artifacts left off.
func applyReentry(opts Options, plansDir string) error {
	if opts.Fresh {
		if err := state.ClearAllArtifacts(plansDir); err != nil {
			return fmt.Errorf("clearing artifacts for fresh start: %w", err)
		}
		writef(opts.Out, "Re-entry: cleared all plan artifacts. Starting at Discovery.\n")
		return nil
	}
	if opts.RerunStage != "" {
		stage, err := state.StageFromString(opts.RerunStage)
		if err != nil {
			return fmt.Errorf("re-run stage: %w", err)
		}
		if err := state.ClearArtifactsFromStage(plansDir, stage); err != nil {
			return fmt.Errorf("clearing artifacts for re-run of %s: %w", stage, err)
		}
		writef(opts.Out, "Re-entry: cleared artifacts from %s onwards.\n", stage)
	}
	return nil
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

// runApprovalGate resolves the approval mode (CLI flag takes precedence over
// config) and runs the gate for the completed stage.
func runApprovalGate(opts Options, cfg *config.Config, stage string) error {
	modeStr := cfg.ApprovalMode
	if opts.ApprovalMode != "" {
		modeStr = opts.ApprovalMode
	}
	if modeStr == "" {
		modeStr = string(approval.ModeAuto)
	}

	mode, err := approval.Parse(modeStr)
	if err != nil {
		return fmt.Errorf("resolving approval mode: %w", err)
	}

	return approval.Gate(mode, stage, opts.Out, opts.In)
}
