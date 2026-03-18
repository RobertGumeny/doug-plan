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

func TestStageFromString(t *testing.T) {
	tests := []struct {
		input   string
		want    Stage
		wantErr bool
	}{
		{"Discovery", StageDiscovery, false},
		{"discovery", StageDiscovery, false},
		{"DISCOVERY", StageDiscovery, false},
		{"Roadmapping", StageRoadmapping, false},
		{"roadmapping", StageRoadmapping, false},
		{"PRD", StagePRD, false},
		{"prd", StagePRD, false},
		{"Tasks", StageTasks, false},
		{"tasks", StageTasks, false},
		{"Complete", StageDiscovery, true},
		{"unknown", StageDiscovery, true},
		{"", StageDiscovery, true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := StageFromString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("StageFromString(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("StageFromString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestClearArtifactsFromStage(t *testing.T) {
	tests := []struct {
		name        string
		present     []string // artifacts present before clear
		clearFrom   Stage
		wantPresent []string // artifacts expected present after clear
		wantAbsent  []string // artifacts expected absent after clear
	}{
		{
			name:        "clear from Discovery removes all",
			present:     []string{"VISION.md", "ROADMAP.md", "PRD.md", "TASKS.md"},
			clearFrom:   StageDiscovery,
			wantPresent: nil,
			wantAbsent:  []string{"VISION.md", "ROADMAP.md", "PRD.md", "TASKS.md"},
		},
		{
			name:        "clear from PRD keeps earlier artifacts",
			present:     []string{"VISION.md", "ROADMAP.md", "PRD.md", "TASKS.md"},
			clearFrom:   StagePRD,
			wantPresent: []string{"VISION.md", "ROADMAP.md"},
			wantAbsent:  []string{"PRD.md", "TASKS.md"},
		},
		{
			name:        "clear from Tasks removes only TASKS.md",
			present:     []string{"VISION.md", "ROADMAP.md", "PRD.md", "TASKS.md"},
			clearFrom:   StageTasks,
			wantPresent: []string{"VISION.md", "ROADMAP.md", "PRD.md"},
			wantAbsent:  []string{"TASKS.md"},
		},
		{
			name:        "clear from Roadmapping with only VISION.md present is no-op for present",
			present:     []string{"VISION.md"},
			clearFrom:   StageRoadmapping,
			wantPresent: []string{"VISION.md"},
			wantAbsent:  []string{"ROADMAP.md"},
		},
		{
			name:        "missing artifacts are silently skipped",
			present:     nil,
			clearFrom:   StageDiscovery,
			wantPresent: nil,
			wantAbsent:  []string{"VISION.md", "ROADMAP.md", "PRD.md", "TASKS.md"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, a := range tt.present {
				if err := os.WriteFile(filepath.Join(dir, a), []byte("stub"), 0644); err != nil {
					t.Fatalf("setup: %v", err)
				}
			}
			if err := ClearArtifactsFromStage(dir, tt.clearFrom); err != nil {
				t.Fatalf("ClearArtifactsFromStage: %v", err)
			}
			for _, a := range tt.wantPresent {
				if _, err := os.Stat(filepath.Join(dir, a)); err != nil {
					t.Errorf("expected %s to be present, got error: %v", a, err)
				}
			}
			for _, a := range tt.wantAbsent {
				if _, err := os.Stat(filepath.Join(dir, a)); !os.IsNotExist(err) {
					t.Errorf("expected %s to be absent, got err: %v", a, err)
				}
			}
		})
	}
}

func TestClearAllArtifacts(t *testing.T) {
	dir := t.TempDir()
	for _, a := range []string{"VISION.md", "ROADMAP.md", "PRD.md", "TASKS.md"} {
		if err := os.WriteFile(filepath.Join(dir, a), []byte("stub"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	if err := ClearAllArtifacts(dir); err != nil {
		t.Fatalf("ClearAllArtifacts: %v", err)
	}
	for _, a := range []string{"VISION.md", "ROADMAP.md", "PRD.md", "TASKS.md"} {
		if _, err := os.Stat(filepath.Join(dir, a)); !os.IsNotExist(err) {
			t.Errorf("expected %s to be absent after ClearAllArtifacts", a)
		}
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
