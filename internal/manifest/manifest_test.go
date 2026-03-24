package manifest

import (
	"strings"
	"testing"
)

// fullFrontmatter is a complete greenfield VISION.md with all optional fields.
const fullFrontmatter = `---
project_mode: "greenfield"
language: "typescript"
runtime: "node"
framework: "nextjs"
package_manager: "pnpm"
build_system: "npm-scripts"
runtime_dependencies:
  - "next"
  - "react"
  - "react-dom"
dev_dependencies:
  - "typescript"
  - "eslint"
bootstrap_constraints:
  - "Deploy on Vercel"
---

# Vision

## Project Name

Acme App
`

// minimalFrontmatter has only the required fields.
const minimalFrontmatter = `---
project_mode: "greenfield"
language: "go"
runtime: "go"
---

# Vision

## Project Name

Minimal Project
`

// noFrontmatter is a legacy VISION.md without any frontmatter.
const noFrontmatter = `# Vision

## Project Name

Legacy App

## Problem Statement

Some content here.
`

func TestParseVisionFrontmatter_Full(t *testing.T) {
	fm, err := ParseVisionFrontmatter(fullFrontmatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm == nil {
		t.Fatal("expected non-nil frontmatter")
	}
	if fm.ProjectMode != "greenfield" {
		t.Errorf("ProjectMode: got %q, want %q", fm.ProjectMode, "greenfield")
	}
	if fm.Language != "typescript" {
		t.Errorf("Language: got %q, want %q", fm.Language, "typescript")
	}
	if fm.Runtime != "node" {
		t.Errorf("Runtime: got %q, want %q", fm.Runtime, "node")
	}
	if fm.Framework != "nextjs" {
		t.Errorf("Framework: got %q, want %q", fm.Framework, "nextjs")
	}
	if fm.PackageManager != "pnpm" {
		t.Errorf("PackageManager: got %q, want %q", fm.PackageManager, "pnpm")
	}
	if fm.BuildSystem != "npm-scripts" {
		t.Errorf("BuildSystem: got %q, want %q", fm.BuildSystem, "npm-scripts")
	}
	if len(fm.RuntimeDependencies) != 3 {
		t.Errorf("RuntimeDependencies: got %d items, want 3", len(fm.RuntimeDependencies))
	}
	if len(fm.DevDependencies) != 2 {
		t.Errorf("DevDependencies: got %d items, want 2", len(fm.DevDependencies))
	}
	if len(fm.BootstrapConstraints) != 1 || fm.BootstrapConstraints[0] != "Deploy on Vercel" {
		t.Errorf("BootstrapConstraints: got %v, want [\"Deploy on Vercel\"]", fm.BootstrapConstraints)
	}
}

func TestParseVisionFrontmatter_Minimal(t *testing.T) {
	fm, err := ParseVisionFrontmatter(minimalFrontmatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm == nil {
		t.Fatal("expected non-nil frontmatter")
	}
	if fm.Language != "go" {
		t.Errorf("Language: got %q, want %q", fm.Language, "go")
	}
	if len(fm.RuntimeDependencies) != 0 {
		t.Errorf("RuntimeDependencies: got %v, want empty", fm.RuntimeDependencies)
	}
}

func TestParseVisionFrontmatter_NoFrontmatter(t *testing.T) {
	fm, err := ParseVisionFrontmatter(noFrontmatter)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fm != nil {
		t.Errorf("expected nil for content without frontmatter, got %+v", fm)
	}
}

func TestParseVisionFrontmatter_Malformed(t *testing.T) {
	malformed := "---\n: invalid: yaml: [\n---\n# Vision\n"
	_, err := ParseVisionFrontmatter(malformed)
	if err == nil {
		t.Fatal("expected error for malformed YAML frontmatter, got nil")
	}
}

func TestValidateScaffoldFrontmatter_Valid(t *testing.T) {
	fm := &VisionFrontmatter{
		ProjectMode: "greenfield",
		Language:    "typescript",
		Runtime:     "node",
	}
	if err := ValidateScaffoldFrontmatter(fm); err != nil {
		t.Errorf("unexpected error for valid frontmatter: %v", err)
	}
}

func TestValidateScaffoldFrontmatter_MissingFields(t *testing.T) {
	tests := []struct {
		name        string
		fm          VisionFrontmatter
		wantMissing []string
	}{
		{
			name:        "missing all required",
			fm:          VisionFrontmatter{},
			wantMissing: []string{"project_mode", "language", "runtime"},
		},
		{
			name:        "missing language and runtime",
			fm:          VisionFrontmatter{ProjectMode: "greenfield"},
			wantMissing: []string{"language", "runtime"},
		},
		{
			name:        "missing runtime only",
			fm:          VisionFrontmatter{ProjectMode: "greenfield", Language: "go"},
			wantMissing: []string{"runtime"},
		},
		{
			name:        "whitespace only counts as missing",
			fm:          VisionFrontmatter{ProjectMode: "  ", Language: "go", Runtime: "go"},
			wantMissing: []string{"project_mode"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScaffoldFrontmatter(&tt.fm)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			for _, field := range tt.wantMissing {
				if !strings.Contains(err.Error(), field) {
					t.Errorf("error %q does not mention missing field %q", err.Error(), field)
				}
			}
		})
	}
}

func TestFromVisionFrontmatter_Full(t *testing.T) {
	fm := &VisionFrontmatter{
		ProjectMode:          "greenfield",
		Language:             "typescript",
		Runtime:              "node",
		Framework:            "nextjs",
		PackageManager:       "pnpm",
		BuildSystem:          "npm-scripts",
		RuntimeDependencies:  []string{"next", "react"},
		DevDependencies:      []string{"typescript"},
		BootstrapConstraints: []string{"Deploy on Vercel"},
	}
	m := FromVisionFrontmatter("Acme App", fm)

	if m.SchemaVersion != 1 {
		t.Errorf("SchemaVersion: got %d, want 1", m.SchemaVersion)
	}
	if m.Project.Name != "Acme App" {
		t.Errorf("Project.Name: got %q, want %q", m.Project.Name, "Acme App")
	}
	if m.Project.Mode != "greenfield" {
		t.Errorf("Project.Mode: got %q, want %q", m.Project.Mode, "greenfield")
	}
	if m.Scaffold.Language != "typescript" {
		t.Errorf("Scaffold.Language: got %q, want %q", m.Scaffold.Language, "typescript")
	}
	if m.Scaffold.Runtime != "node" {
		t.Errorf("Scaffold.Runtime: got %q, want %q", m.Scaffold.Runtime, "node")
	}
	if m.Scaffold.Framework != "nextjs" {
		t.Errorf("Scaffold.Framework: got %q, want %q", m.Scaffold.Framework, "nextjs")
	}
	if len(m.Dependencies.Runtime) != 2 {
		t.Errorf("Dependencies.Runtime: got %d items, want 2", len(m.Dependencies.Runtime))
	}
	if len(m.Dependencies.Development) != 1 {
		t.Errorf("Dependencies.Development: got %d items, want 1", len(m.Dependencies.Development))
	}
	if len(m.Constraints) != 1 || m.Constraints[0] != "Deploy on Vercel" {
		t.Errorf("Constraints: got %v", m.Constraints)
	}
}

func TestFromVisionFrontmatter_Minimal(t *testing.T) {
	fm := &VisionFrontmatter{
		ProjectMode: "greenfield",
		Language:    "go",
		Runtime:     "go",
	}
	m := FromVisionFrontmatter("Minimal", fm)

	if m.SchemaVersion != 1 {
		t.Errorf("SchemaVersion: got %d, want 1", m.SchemaVersion)
	}
	if m.Scaffold.Framework != "" {
		t.Errorf("Scaffold.Framework: expected empty, got %q", m.Scaffold.Framework)
	}
	if len(m.Dependencies.Runtime) != 0 {
		t.Errorf("Dependencies.Runtime: expected empty, got %v", m.Dependencies.Runtime)
	}
	if len(m.Constraints) != 0 {
		t.Errorf("Constraints: expected empty, got %v", m.Constraints)
	}
}

func TestMarshal_ContainsRequiredFields(t *testing.T) {
	m := &Manifest{
		SchemaVersion: 1,
		Project:       Project{Name: "Acme App", Mode: "greenfield"},
		Scaffold:      Scaffold{Language: "typescript", Runtime: "node", Framework: "nextjs"},
		Dependencies: Dependencies{
			Runtime:     []string{"next", "react", "react-dom"},
			Development: []string{"typescript", "eslint"},
		},
		Constraints: []string{"Deploy on Vercel"},
	}
	data, err := Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	yaml := string(data)

	for _, expected := range []string{
		"schema_version: 1",
		"name: Acme App",
		"mode: greenfield",
		"language: typescript",
		"runtime: node",
		"framework: nextjs",
		"- next",
		"- typescript",
		"- Deploy on Vercel",
	} {
		if !strings.Contains(yaml, expected) {
			t.Errorf("Marshal output missing %q\nGot:\n%s", expected, yaml)
		}
	}
}

func TestMarshal_OmitsEmptyOptionalFields(t *testing.T) {
	m := &Manifest{
		SchemaVersion: 1,
		Project:       Project{Name: "Minimal", Mode: "greenfield"},
		Scaffold:      Scaffold{Language: "go", Runtime: "go"},
	}
	data, err := Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	yaml := string(data)

	for _, absent := range []string{"framework", "package_manager", "build_system", "dependencies", "constraints"} {
		if strings.Contains(yaml, absent) {
			t.Errorf("Marshal output should omit empty field %q\nGot:\n%s", absent, yaml)
		}
	}
}
