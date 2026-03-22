package state

import (
	"testing"
)

func TestValidateVision(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid: has non-heading content",
			content: "# Vision\n\n## Problem Statement\n\nSome real content here.\n",
			wantErr: false,
		},
		{
			name:    "valid: content not under a heading",
			content: "Some content without headings.\n",
			wantErr: false,
		},
		{
			name:    "invalid: only headings and blank lines (unfilled shell)",
			content: "# Vision\n\n## Project Name\n\n## Problem Statement\n\n## Goals\n",
			wantErr: true,
		},
		{
			name:    "invalid: empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "invalid: only blank lines",
			content: "\n\n\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateVision(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateVision() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateRoadmap(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid: non-empty project field",
			content: "---\nproject: \"MyProject\"\nsource: VISION.md\n---\n\n# Roadmap\n",
			wantErr: false,
		},
		{
			name:    "valid: no frontmatter at all",
			content: "# Roadmap\n\nSome content.\n",
			wantErr: false,
		},
		{
			name:    "invalid: empty project field (unfilled shell)",
			content: "---\nproject: \"\"\nsource: VISION.md\n---\n\n# Roadmap\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRoadmap(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRoadmap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateDefinition(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid: non-empty id field",
			content: "---\nid: \"EPIC-1\"\nname: \"Core Feature\"\n---\n\n# Definition\n",
			wantErr: false,
		},
		{
			name:    "valid: completion marker without frontmatter",
			content: "# Definition Complete\n\nAll epics defined.\n",
			wantErr: false,
		},
		{
			name:    "invalid: empty id field (unfilled shell)",
			content: "---\nid: \"\"\nname: \"\"\n---\n\n# Definition\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDefinition(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDefinition() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePRD(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid: starts with top-level heading",
			content: "# PRD — EPIC-1: Core Feature\n\n## Overview\n\nContent.\n",
			wantErr: false,
		},
		{
			name:    "valid: handoff complete heading",
			content: "# Handoff Complete\n\nAll epics have been defined.\n",
			wantErr: false,
		},
		{
			name:    "invalid: does not start with heading",
			content: "Some preamble text.\n\n# PRD\n",
			wantErr: true,
		},
		{
			name:    "invalid: empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "invalid: starts with ## instead of #",
			content: "## Section\n\nContent.\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePRD(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePRD() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateTasksYAML(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid: starts with epics key",
			content: "epics:\n  - id: \"EPIC-1\"\n    name: \"Core Feature\"\n",
			wantErr: false,
		},
		{
			name:    "valid: epics key with leading whitespace trimmed",
			content: "\nepics:\n  - id: \"EPIC-1\"\n",
			wantErr: false,
		},
		{
			name:    "invalid: starts with wrong key",
			content: "epic:\n  id: \"EPIC-1\"\n",
			wantErr: true,
		},
		{
			name:    "invalid: empty file",
			content: "",
			wantErr: true,
		},
		{
			name:    "invalid: not yaml",
			content: "# some markdown\n",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTasksYAML(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTasksYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
