package handoff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const fullDefinition = `---
id: "EPIC-1"
name: "Core Habit Tracking"
---

# Epic Definition: EPIC-1 — Core Habit Tracking

**Generated**: 2026-03-20
**Epic ID**: EPIC-1
**Source**: ROADMAP.md

---

## Overview

Build the foundational habit-creation and daily check-in flow.

---

## Scope

### In-Scope

- Habit creation and deletion
- Daily check-in with single-tap interaction

### Out-of-Scope

- Social features

## Goals

- Allow users to create up to ten daily habits.
- Persist data locally on device.

## Non-Goals

- Backend integration in v1.

## Background

This epic establishes the data model and UI skeleton.

## Success Criteria

- Users can create a habit in under three taps.
- Data persists across app restarts.

## Deliverables

- Habit data model
- Check-in UI component

## Tasks

### EPIC-1-001: Implement habit data model

**Type**: feature
**Description**: Define the Habit entity with id, name, and completion records.

**Acceptance Criteria**:
- Habit entity has id, name, and daily completion records
- Data is persisted to local storage
- Unit tests cover CRUD operations

### EPIC-1-002: Build check-in UI

**Type**: feature
**Description**: Single-tap check-in screen listing all habits for today.

**Acceptance Criteria**:
- Check-in screen renders all habits
- Single tap marks a habit as done for the day
`

func TestParseDefinition_Full(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "DEFINITION.md")
	if err := os.WriteFile(path, []byte(fullDefinition), 0o644); err != nil {
		t.Fatal(err)
	}

	def, err := ParseDefinition(path, "EPIC-1")
	if err != nil {
		t.Fatalf("ParseDefinition: %v", err)
	}

	if def.ID != "EPIC-1" {
		t.Errorf("ID: got %q, want %q", def.ID, "EPIC-1")
	}
	if def.Name != "Core Habit Tracking" {
		t.Errorf("Name: got %q, want %q", def.Name, "Core Habit Tracking")
	}
	if !strings.Contains(def.Overview, "foundational habit-creation") {
		t.Errorf("Overview: unexpected content: %q", def.Overview)
	}
	if !strings.Contains(def.Goals, "ten daily habits") {
		t.Errorf("Goals: unexpected content: %q", def.Goals)
	}
	if !strings.Contains(def.SuccessCriteria, "three taps") {
		t.Errorf("SuccessCriteria: unexpected content: %q", def.SuccessCriteria)
	}
	if len(def.Tasks) != 2 {
		t.Fatalf("Tasks: got %d, want 2", len(def.Tasks))
	}
	if def.Tasks[0].ID != "EPIC-1-001" {
		t.Errorf("Tasks[0].ID: got %q, want %q", def.Tasks[0].ID, "EPIC-1-001")
	}
	if def.Tasks[0].Type != "feature" {
		t.Errorf("Tasks[0].Type: got %q, want %q", def.Tasks[0].Type, "feature")
	}
	if len(def.Tasks[0].Acceptance) != 3 {
		t.Errorf("Tasks[0].Acceptance: got %d items, want 3", len(def.Tasks[0].Acceptance))
	}
	if def.Tasks[1].ID != "EPIC-1-002" {
		t.Errorf("Tasks[1].ID: got %q, want %q", def.Tasks[1].ID, "EPIC-1-002")
	}
}

func TestParseDefinition_Minimal(t *testing.T) {
	// Fakeagent-style stub — just a heading, no frontmatter.
	stub := "# Definition: EPIC-2\n"
	dir := t.TempDir()
	path := filepath.Join(dir, "DEFINITION.md")
	if err := os.WriteFile(path, []byte(stub), 0o644); err != nil {
		t.Fatal(err)
	}

	def, err := ParseDefinition(path, "EPIC-2")
	if err != nil {
		t.Fatalf("ParseDefinition: %v", err)
	}
	if def.ID != "EPIC-2" {
		t.Errorf("ID: got %q, want %q", def.ID, "EPIC-2")
	}
	if def.Name != "EPIC-2" {
		t.Errorf("Name: got %q, want %q", def.Name, "EPIC-2")
	}
	if len(def.Tasks) != 0 {
		t.Errorf("Tasks: got %d, want 0", len(def.Tasks))
	}
}

func TestRenderPRD_ContainsRequiredSections(t *testing.T) {
	def := &EpicDefinition{
		ID:              "EPIC-1",
		Name:            "Core Habit Tracking",
		Overview:        "Build the foundational flow.",
		Goals:           "- Goal A",
		SuccessCriteria: "- Passes unit tests",
		Tasks: []TaskDef{
			{ID: "EPIC-1-001", Name: "Task One", Type: "feature", Description: "Do something.", Acceptance: []string{"AC 1", "AC 2"}},
		},
	}

	prd := RenderPRD(def)

	for _, expected := range []string{
		"# PRD — EPIC-1: Core Habit Tracking",
		"## Epic",
		"- id: EPIC-1",
		"## Overview",
		"Build the foundational flow.",
		"## Goals",
		"## Acceptance Criteria",
		"- AC 1",
		"- AC 2",
		"## Notes for Agents",
	} {
		if !strings.Contains(prd, expected) {
			t.Errorf("RenderPRD missing %q", expected)
		}
	}
}

func TestRenderTasksYAML_WithTasks(t *testing.T) {
	def := &EpicDefinition{
		ID:   "EPIC-1",
		Name: "Core Habit Tracking",
		Tasks: []TaskDef{
			{
				ID:          "EPIC-1-001",
				Name:        "Task One",
				Type:        "feature",
				Description: "Do something concrete.",
				Acceptance:  []string{"Criterion A", "Criterion B"},
			},
		},
	}

	yaml := RenderTasksYAML(def)

	for _, expected := range []string{
		`epic:`,
		`id: "EPIC-1"`,
		`name: "Core Habit Tracking"`,
		`- id: "EPIC-1-001"`,
		`type: "feature"`,
		`status: "TODO"`,
		`description: "Do something concrete."`,
		`acceptance_criteria:`,
		`- "Criterion A"`,
	} {
		if !strings.Contains(yaml, expected) {
			t.Errorf("RenderTasksYAML missing %q\nGot:\n%s", expected, yaml)
		}
	}
}

func TestRenderTasksYAML_NoTasks(t *testing.T) {
	def := &EpicDefinition{ID: "EPIC-2", Name: "Stub Epic"}
	yaml := RenderTasksYAML(def)
	if !strings.Contains(yaml, "tasks: []") {
		t.Errorf("expected 'tasks: []' for empty task list, got:\n%s", yaml)
	}
}

func TestExecute_ProducesAllArtifacts(t *testing.T) {
	dir := t.TempDir()
	planDir := filepath.Join(dir, ".doug", "plan")
	epicsDir := filepath.Join(planDir, "epics")

	epicIDs := []string{"EPIC-1", "EPIC-2"}
	for _, id := range epicIDs {
		epicDir := filepath.Join(epicsDir, id)
		if err := os.MkdirAll(epicDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := "# Definition: " + id + "\n"
		if err := os.WriteFile(filepath.Join(epicDir, "DEFINITION.md"), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := Execute(planDir); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	// Per-epic artifacts.
	for _, id := range epicIDs {
		epicDir := filepath.Join(epicsDir, id)
		for _, f := range []string{"PRD.md", "tasks.yaml"} {
			path := filepath.Join(epicDir, f)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("missing per-epic %s for %s", f, id)
			}
		}
	}

	// Global artifacts.
	for _, f := range []string{"PRD.md", "tasks.yaml"} {
		if _, err := os.Stat(filepath.Join(planDir, f)); os.IsNotExist(err) {
			t.Errorf("missing global %s", f)
		}
	}

	// Global PRD.md should reference both epics.
	data, err := os.ReadFile(filepath.Join(planDir, "PRD.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range epicIDs {
		if !strings.Contains(string(data), id) {
			t.Errorf("global PRD.md does not reference %s", id)
		}
	}
}

func TestExecute_NoEpicsDir(t *testing.T) {
	dir := t.TempDir()
	planDir := filepath.Join(dir, ".doug", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Execute(planDir); err == nil {
		t.Fatal("expected error for missing epics dir, got nil")
	}
}

func TestExecute_NoDefinitions(t *testing.T) {
	dir := t.TempDir()
	planDir := filepath.Join(dir, ".doug", "plan")
	epicsDir := filepath.Join(planDir, "epics")
	// Create epics dir but no per-epic dirs with DEFINITION.md.
	if err := os.MkdirAll(epicsDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := Execute(planDir); err == nil {
		t.Fatal("expected error when no definitions found, got nil")
	}
}
