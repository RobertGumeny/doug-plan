package orchestrator

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/robertgumeny/doug-plan/internal/agent"
	"github.com/robertgumeny/doug-plan/internal/approval"
	"github.com/robertgumeny/doug-plan/internal/config"
	"github.com/robertgumeny/doug-plan/internal/state"
)

// Options holds the runtime configuration for an orchestrator run.
type Options struct {
	ProjectRoot  string
	Out          io.Writer
	In           io.Reader
	ApprovalMode string // overrides config when non-empty; must be "auto", "soft", or "hard"
}

// Run infers the current pipeline position from artifacts on disk,
// writes an ACTIVE_STEP.md briefing, invokes the configured agent,
// parses the result, dispatches by outcome, and archives the step file.
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

	cfg, err := config.Load(opts.ProjectRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	args, err := cfg.AgentCommand()
	if err != nil {
		return fmt.Errorf("resolving agent command: %w", err)
	}

	fmt.Fprintf(opts.Out, "Invoking agent: %s\n", args[0])
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
		fmt.Fprintf(opts.Out, "Step %s completed successfully.\n", stage)
		if err := runApprovalGate(opts, cfg, stage.String()); err != nil {
			if errors.Is(err, approval.ErrSkipped) {
				return nil
			}
			return err
		}
	case agent.OutcomeFailure:
		return fmt.Errorf("step %s failed: agent reported FAILURE", stage)
	case agent.OutcomeRetry:
		fmt.Fprintf(opts.Out, "Step %s requesting retry.\n", stage)
	}

	return nil
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
