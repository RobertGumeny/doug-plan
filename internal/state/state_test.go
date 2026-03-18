package state

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInferStage(t *testing.T) {
	tests := []struct {
		name      string
		artifacts []string // files to create in the temp plans dir
		want      Stage
	}{
		{
			name:      "no artifacts → Discovery",
			artifacts: nil,
			want:      StageDiscovery,
		},
		{
			name:      "VISION.md only → Roadmapping",
			artifacts: []string{"VISION.md"},
			want:      StageRoadmapping,
		},
		{
			name:      "VISION.md + ROADMAP.md → PRD",
			artifacts: []string{"VISION.md", "ROADMAP.md"},
			want:      StagePRD,
		},
		{
			name:      "VISION.md + ROADMAP.md + PRD.md → Tasks",
			artifacts: []string{"VISION.md", "ROADMAP.md", "PRD.md"},
			want:      StageTasks,
		},
		{
			name:      "all artifacts → Complete",
			artifacts: []string{"VISION.md", "ROADMAP.md", "PRD.md", "TASKS.md"},
			want:      StageComplete,
		},
		{
			name:      "gap in sequence uses first missing artifact",
			artifacts: []string{"ROADMAP.md"}, // VISION.md absent → still Discovery
			want:      StageDiscovery,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, artifact := range tt.artifacts {
				if err := os.WriteFile(filepath.Join(dir, artifact), []byte("stub"), 0644); err != nil {
					t.Fatalf("setup: writing %s: %v", artifact, err)
				}
			}

			got, err := InferStage(dir)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("InferStage() = %v (%d), want %v (%d)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestStageString(t *testing.T) {
	tests := []struct {
		stage Stage
		want  string
	}{
		{StageDiscovery, "Discovery"},
		{StageRoadmapping, "Roadmapping"},
		{StagePRD, "PRD"},
		{StageTasks, "Tasks"},
		{StageComplete, "Complete"},
	}
	for _, tt := range tests {
		if got := tt.stage.String(); got != tt.want {
			t.Errorf("Stage(%d).String() = %q, want %q", tt.stage, got, tt.want)
		}
	}
}
