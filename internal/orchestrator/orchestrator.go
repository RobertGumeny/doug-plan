package orchestrator

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/robertgumeny/doug-plan/internal/agent"
	"github.com/robertgumeny/doug-plan/internal/approval"
	"github.com/robertgumeny/doug-plan/internal/config"
	"github.com/robertgumeny/doug-plan/internal/handoff"
	"github.com/robertgumeny/doug-plan/internal/layout"
	"github.com/robertgumeny/doug-plan/internal/state"
)

var (
	invokeAgent      = agent.Invoke
	browserGateFunc  = approval.BrowserGate
	terminalGateFunc = approval.Gate
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

	// StagePRD is handled deterministically — no agent invocation.
	if stage == state.StagePRD {
		return runHandoff(opts, plansDir)
	}

	if err := agent.MaterializeArtifact(opts.ProjectRoot, stage); err != nil {
		return fmt.Errorf("materializing artifact shell: %w", err)
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
	if err := invokeAgent(opts.ProjectRoot, args); err != nil {
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
		if err := runApprovalGate(opts, cfg, stage, plansDir); err != nil {
			if errors.Is(err, approval.ErrSkipped) {
				return nil
			}
			return err
		}
	case agent.OutcomeFailure:
		return fmt.Errorf("step %s failed: agent reported FAILURE", stage)
	case agent.OutcomeRetry:
		writef(opts.Out, "Step %s requesting retry.\n", stage)
		// Remove the artifact shell so InferStage re-enters this stage on the
		// next run. On a RETRY the agent has not finished the stage, so any file
		// at the artifact path is either the host-created shell or partial work
		// that should be regenerated. Silently ignore not-exist errors.
		if artifactFile := state.ArtifactFile(stage); artifactFile != "" {
			_ = os.Remove(filepath.Join(plansDir, artifactFile))
		}
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

// runHandoff executes the deterministic handoff renderer for StagePRD.
// It renders PRD.md and tasks.yaml from per-epic DEFINITION.md files and then
// runs the approval gate so the user can review both before the stage advances.
func runHandoff(opts Options, plansDir string) error {
	writef(opts.Out, "Pipeline entry point: %s\n", state.StagePRD)

	if err := handoff.Execute(plansDir); err != nil {
		return fmt.Errorf("deterministic handoff: %w", err)
	}

	writef(opts.Out, "Step %s completed successfully.\n", state.StagePRD)

	cfg, err := config.Load(opts.ProjectRoot)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	if err := runApprovalGate(opts, cfg, state.StagePRD, plansDir); err != nil {
		if errors.Is(err, approval.ErrSkipped) {
			return nil
		}
		return err
	}
	return nil
}

// runApprovalGate resolves the approval mode (CLI flag takes precedence over
// config) and runs the gate for the completed stage.
// Hard mode opens the browser review UI; auto and soft use terminal gates.
func runApprovalGate(opts Options, cfg *config.Config, stage state.Stage, plansDir string) error {
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

	if mode == approval.ModeHard {
		artifactFile := state.ArtifactFile(stage)
		if artifactFile != "" {
			primaryPath := filepath.Join(plansDir, artifactFile)
			secondaryPath := ""
			if stage == state.StagePRD {
				candidate := filepath.Join(plansDir, "tasks.yaml")
				if _, statErr := os.Stat(candidate); statErr == nil {
					secondaryPath = candidate
				}
			}
			return browserGateFunc(primaryPath, secondaryPath, stage.String(), opts.Out)
		}
	}

	return terminalGateFunc(mode, stage.String(), opts.Out, opts.In)
}
