package manifest

import (
	"os"
	"path/filepath"
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

// --- Sync tests ---

// makePlanDir creates a temp project root with a .doug/plan directory,
// writes visionContent to VISION.md, and returns the project root path.
func makePlanDir(t *testing.T, visionContent string) string {
	t.Helper()
	root := t.TempDir()
	planDir := filepath.Join(root, ".doug", "plan")
	if err := os.MkdirAll(planDir, 0o755); err != nil {
		t.Fatalf("mkdir planDir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "VISION.md"), []byte(visionContent), 0o644); err != nil {
		t.Fatalf("write VISION.md: %v", err)
	}
	return root
}

func TestSync_WritesManifestForGreenfield(t *testing.T) {
	root := makePlanDir(t, fullFrontmatter)
	if err := Sync(root); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	manifestPath := filepath.Join(root, ".doug", "plan", "manifest.yaml")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("reading manifest: %v", err)
	}
	content := string(data)
	for _, want := range []string{
		"schema_version: 1",
		"mode: greenfield",
		"language: typescript",
		"runtime: node",
		"- next",
		"- Deploy on Vercel",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("manifest missing %q\nGot:\n%s", want, content)
		}
	}
}

func TestSync_ExtractsProjectName(t *testing.T) {
	root := makePlanDir(t, fullFrontmatter)
	if err := Sync(root); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	data, err := os.ReadFile(filepath.Join(root, ".doug", "plan", "manifest.yaml"))
	if err != nil {
		t.Fatalf("reading manifest: %v", err)
	}
	if !strings.Contains(string(data), "name: Acme App") {
		t.Errorf("manifest missing project name\nGot:\n%s", string(data))
	}
}

func TestSync_RemovesManifestForNonGreenfield(t *testing.T) {
	existingVision := `---
project_mode: "existing"
language: "go"
runtime: "go"
---

# Vision

## Project Name

Legacy App
`
	root := makePlanDir(t, existingVision)
	manifestPath := filepath.Join(root, ".doug", "plan", "manifest.yaml")
	// Pre-create a stale manifest.
	if err := os.WriteFile(manifestPath, []byte("stale: true\n"), 0o644); err != nil {
		t.Fatalf("writing stale manifest: %v", err)
	}

	if err := Sync(root); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Error("expected manifest to be removed for non-greenfield project, but it still exists")
	}
}

func TestSync_RemovesManifestWhenNoFrontmatter(t *testing.T) {
	root := makePlanDir(t, noFrontmatter)
	manifestPath := filepath.Join(root, ".doug", "plan", "manifest.yaml")
	if err := os.WriteFile(manifestPath, []byte("stale: true\n"), 0o644); err != nil {
		t.Fatalf("writing stale manifest: %v", err)
	}

	if err := Sync(root); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	if _, err := os.Stat(manifestPath); !os.IsNotExist(err) {
		t.Error("expected manifest to be removed when no frontmatter, but it still exists")
	}
}

func TestSync_SucceedsWhenNoStaleManifest(t *testing.T) {
	root := makePlanDir(t, noFrontmatter)
	// No manifest exists — Sync should not error.
	if err := Sync(root); err != nil {
		t.Fatalf("Sync: %v", err)
	}
}

func TestSync_ErrorsOnMissingRequiredFields(t *testing.T) {
	missingFields := `---
project_mode: "greenfield"
---

# Vision

## Project Name

Incomplete App
`
	root := makePlanDir(t, missingFields)
	err := Sync(root)
	if err == nil {
		t.Fatal("expected error for missing required fields, got nil")
	}
	if !strings.Contains(err.Error(), "language") || !strings.Contains(err.Error(), "runtime") {
		t.Errorf("error should mention missing fields, got: %v", err)
	}
	// No partial manifest should be written.
	manifestPath := filepath.Join(root, ".doug", "plan", "manifest.yaml")
	if _, statErr := os.Stat(manifestPath); !os.IsNotExist(statErr) {
		t.Error("partial manifest must not be written on validation failure")
	}
}

func TestSync_ErrorsOnMalformedFrontmatter(t *testing.T) {
	root := makePlanDir(t, "---\n: invalid: yaml: [\n---\n# Vision\n")
	if err := Sync(root); err == nil {
		t.Fatal("expected error for malformed frontmatter, got nil")
	}
}

func TestSync_AtomicWrite_NoTmpLeftOnSuccess(t *testing.T) {
	root := makePlanDir(t, fullFrontmatter)
	if err := Sync(root); err != nil {
		t.Fatalf("Sync: %v", err)
	}
	tmp := filepath.Join(root, ".doug", "plan", "manifest.yaml.tmp")
	if _, err := os.Stat(tmp); !os.IsNotExist(err) {
		t.Error("tmp file should not remain after successful Sync")
	}
}

func TestExtractProjectName(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name:    "name from section",
			content: "# Vision\n\n## Project Name\n\nAcme App\n\n## Goals\n\nSome goals\n",
			want:    "Acme App",
		},
		{
			name:    "name from section with frontmatter",
			content: fullFrontmatter,
			want:    "Acme App",
		},
		{
			name:    "no project name section",
			content: "# Vision\n\n## Goals\n\nSome goals\n",
			want:    "",
		},
		{
			name:    "empty project name section",
			content: "# Vision\n\n## Project Name\n\n## Goals\n",
			want:    "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractProjectName(tt.content)
			if got != tt.want {
				t.Errorf("extractProjectName: got %q, want %q", got, tt.want)
			}
		})
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
