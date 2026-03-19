package state

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Stage represents a pipeline step, ordered from first to last.
type Stage int

const (
	StageDiscovery   Stage = iota // produces VISION.md
	StageRoadmapping              // produces ROADMAP.md
	StageScoping                  // produces SCOPED.md
	StagePRD                      // produces PRD.md
	StageTasks                    // produces TASKS.md
	StageComplete
)

// String returns the display name for a Stage.
func (s Stage) String() string {
	switch s {
	case StageDiscovery:
		return "Discovery"
	case StageRoadmapping:
		return "Roadmapping"
	case StageScoping:
		return "Scoping"
	case StagePRD:
		return "PRD"
	case StageTasks:
		return "Tasks"
	case StageComplete:
		return "Complete"
	default:
		return "Unknown"
	}
}

// pipelineStep pairs a stage with the artifact it produces in .doug/plan/.
type pipelineStep struct {
	stage    Stage
	artifact string
}

// pipeline defines the ordered stages and their output artifacts.
// A stage is considered done when its artifact is present on disk.
// The entry point is the first stage whose artifact is absent.
var pipeline = []pipelineStep{
	{StageDiscovery, "VISION.md"},
	{StageRoadmapping, "ROADMAP.md"},
	{StageScoping, "SCOPED.md"},
	{StagePRD, "PRD.md"},
	{StageTasks, "TASKS.md"},
}

// StageFromString parses a stage name (case-insensitive) and returns the
// corresponding Stage value, or an error if the name is not recognized.
func StageFromString(name string) (Stage, error) {
	switch strings.ToLower(name) {
	case "discovery":
		return StageDiscovery, nil
	case "roadmapping":
		return StageRoadmapping, nil
	case "scoping":
		return StageScoping, nil
	case "prd":
		return StagePRD, nil
	case "tasks":
		return StageTasks, nil
	default:
		return StageDiscovery, fmt.Errorf("unknown stage %q: must be one of Discovery, Roadmapping, Scoping, PRD, Tasks", name)
	}
}

// ClearArtifactsFromStage removes the artifact for the given stage and all
// subsequent stages from plansDir. Missing artifacts are silently skipped.
func ClearArtifactsFromStage(plansDir string, stage Stage) error {
	for _, step := range pipeline {
		if step.stage < stage {
			continue
		}
		path := filepath.Join(plansDir, step.artifact)
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing artifact %s: %w", step.artifact, err)
		}
	}
	return nil
}

// ClearAllArtifacts removes all pipeline artifacts from plansDir.
// Missing artifacts are silently skipped.
func ClearAllArtifacts(plansDir string) error {
	return ClearArtifactsFromStage(plansDir, StageDiscovery)
}

// InferStage inspects plansDir and returns the first stage whose artifact is
// absent. If all artifacts are present, StageComplete is returned.
// No state file is read or written.
func InferStage(plansDir string) (Stage, error) {
	for _, step := range pipeline {
		path := filepath.Join(plansDir, step.artifact)
		_, err := os.Stat(path)
		if os.IsNotExist(err) {
			return step.stage, nil
		}
		if err != nil {
			return StageDiscovery, fmt.Errorf("checking artifact %s: %w", step.artifact, err)
		}
	}
	return StageComplete, nil
}
