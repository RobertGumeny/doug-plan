package state

import (
	"os"
	"path/filepath"
	"testing"
)

// validArtifactContent maps each managed artifact filename to minimal content
// that satisfies its validator. Tests that need a "present and valid" artifact
// use this map so InferStage does not treat the file as incomplete.
var validArtifactContent = map[string]string{
	"VISION.md":     "# Vision\n\nContent here.\n",
	"ROADMAP.md":    "---\nproject: \"MyProject\"\nsource: VISION.md\n---\n\n# Roadmap\n",
	"DEFINITION.md": "# Definition Complete\n\nAll epics defined.\n",
	"PRD.md":        "# PRD\n\nContent here.\n",
	"tasks.yaml":    "epics:\n  - id: \"EPIC-1\"\n",
}

func TestInferStage(t *testing.T) {
	tests := []struct {
		name      string
		artifacts []string // files to create with valid content
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
			name:      "VISION.md + ROADMAP.md → Definition",
			artifacts: []string{"VISION.md", "ROADMAP.md"},
			want:      StageDefinition,
		},
		{
			name:      "VISION.md + ROADMAP.md + DEFINITION.md → PRD",
			artifacts: []string{"VISION.md", "ROADMAP.md", "DEFINITION.md"},
			want:      StagePRD,
		},
		{
			name:      "VISION.md + ROADMAP.md + DEFINITION.md + PRD.md → Tasks",
			artifacts: []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md"},
			want:      StageTasks,
		},
		{
			name:      "all artifacts → Complete",
			artifacts: []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md", "tasks.yaml"},
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
				content, ok := validArtifactContent[artifact]
				if !ok {
					content = "stub"
				}
				if err := os.WriteFile(filepath.Join(dir, artifact), []byte(content), 0644); err != nil {
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

func TestInferStage_InvalidArtifacts(t *testing.T) {
	// invalidArtifactContent maps each artifact to content that fails its
	// validator (i.e. the unfilled shell written by MaterializeArtifact).
	invalidArtifactContent := map[string]string{
		"VISION.md":     "# Vision\n\n## Problem Statement\n\n## Goals\n",
		"ROADMAP.md":    "---\nproject: \"\"\nsource: VISION.md\n---\n\n# Roadmap\n",
		"DEFINITION.md": "---\nid: \"\"\nname: \"\"\n---\n\n# Definition\n",
		"PRD.md":        "not a heading\n",
		"tasks.yaml":    "# not yaml\n",
	}

	tests := []struct {
		name            string
		validArtifacts  []string // present with valid content
		invalidArtifact string   // present with invalid (shell) content
		want            Stage
	}{
		{
			name:            "invalid VISION.md → re-enters Discovery",
			validArtifacts:  nil,
			invalidArtifact: "VISION.md",
			want:            StageDiscovery,
		},
		{
			name:            "valid VISION.md + invalid ROADMAP.md → re-enters Roadmapping",
			validArtifacts:  []string{"VISION.md"},
			invalidArtifact: "ROADMAP.md",
			want:            StageRoadmapping,
		},
		{
			name:            "valid VISION.md + ROADMAP.md + invalid DEFINITION.md → re-enters Definition",
			validArtifacts:  []string{"VISION.md", "ROADMAP.md"},
			invalidArtifact: "DEFINITION.md",
			want:            StageDefinition,
		},
		{
			name:            "all prior valid + invalid PRD.md → re-enters PRD",
			validArtifacts:  []string{"VISION.md", "ROADMAP.md", "DEFINITION.md"},
			invalidArtifact: "PRD.md",
			want:            StagePRD,
		},
		{
			name:            "all prior valid + invalid tasks.yaml → re-enters Tasks",
			validArtifacts:  []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md"},
			invalidArtifact: "tasks.yaml",
			want:            StageTasks,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			for _, artifact := range tt.validArtifacts {
				content := validArtifactContent[artifact]
				if err := os.WriteFile(filepath.Join(dir, artifact), []byte(content), 0644); err != nil {
					t.Fatalf("setup: writing valid %s: %v", artifact, err)
				}
			}
			invalidContent := invalidArtifactContent[tt.invalidArtifact]
			if err := os.WriteFile(filepath.Join(dir, tt.invalidArtifact), []byte(invalidContent), 0644); err != nil {
				t.Fatalf("setup: writing invalid %s: %v", tt.invalidArtifact, err)
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
		{"Definition", StageDefinition, false},
		{"definition", StageDefinition, false},
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
			present:     []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md", "tasks.yaml"},
			clearFrom:   StageDiscovery,
			wantPresent: nil,
			wantAbsent:  []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md", "tasks.yaml"},
		},
		{
			name:        "clear from Definition keeps earlier artifacts",
			present:     []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md", "tasks.yaml"},
			clearFrom:   StageDefinition,
			wantPresent: []string{"VISION.md", "ROADMAP.md"},
			wantAbsent:  []string{"DEFINITION.md", "PRD.md", "tasks.yaml"},
		},
		{
			name:        "clear from PRD keeps earlier artifacts",
			present:     []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md", "tasks.yaml"},
			clearFrom:   StagePRD,
			wantPresent: []string{"VISION.md", "ROADMAP.md", "DEFINITION.md"},
			wantAbsent:  []string{"PRD.md", "tasks.yaml"},
		},
		{
			name:        "clear from Tasks removes only tasks.yaml",
			present:     []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md", "tasks.yaml"},
			clearFrom:   StageTasks,
			wantPresent: []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md"},
			wantAbsent:  []string{"tasks.yaml"},
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
			wantAbsent:  []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md", "tasks.yaml"},
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
	for _, a := range []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md", "tasks.yaml"} {
		if err := os.WriteFile(filepath.Join(dir, a), []byte("stub"), 0644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	if err := ClearAllArtifacts(dir); err != nil {
		t.Fatalf("ClearAllArtifacts: %v", err)
	}
	for _, a := range []string{"VISION.md", "ROADMAP.md", "DEFINITION.md", "PRD.md", "tasks.yaml"} {
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
		{StageDefinition, "Definition"},
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
