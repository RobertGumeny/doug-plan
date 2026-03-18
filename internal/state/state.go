package state

import (
	"fmt"
	"os"
	"path/filepath"
)

// Stage represents a pipeline step, ordered from first to last.
type Stage int

const (
	StageDiscovery   Stage = iota // produces VISION.md
	StageRoadmapping              // produces ROADMAP.md
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

// pipelineStep pairs a stage with the artifact it produces in .doug/plans/.
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
	{StagePRD, "PRD.md"},
	{StageTasks, "TASKS.md"},
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
