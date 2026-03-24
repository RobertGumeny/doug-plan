// Package manifest provides schema types and helpers for the manifest.yaml
// planning artifact produced from VISION.md greenfield frontmatter.
//
// Schema v1 required fields: schema_version, project.name, project.mode,
// scaffold.language, scaffold.runtime.
//
// All other scaffold and dependency fields are optional.
package manifest

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/robertgumeny/doug-plan/internal/layout"
	"gopkg.in/yaml.v3"
)

// schemaVersion is the current manifest schema version.
const schemaVersion = 1

// frontmatterRE matches a YAML frontmatter block at the very start of a document.
var frontmatterRE = regexp.MustCompile(`(?s)^---\n(.+?)\n---\n`)

// VisionFrontmatter holds the YAML frontmatter fields that a greenfield
// VISION.md must supply to drive manifest generation.
type VisionFrontmatter struct {
	ProjectMode          string   `yaml:"project_mode"`
	Language             string   `yaml:"language"`
	Runtime              string   `yaml:"runtime"`
	Framework            string   `yaml:"framework"`
	PackageManager       string   `yaml:"package_manager"`
	BuildSystem          string   `yaml:"build_system"`
	RuntimeDependencies  []string `yaml:"runtime_dependencies"`
	DevDependencies      []string `yaml:"dev_dependencies"`
	BootstrapConstraints []string `yaml:"bootstrap_constraints"`
}

// Manifest represents the manifest.yaml schema v1 artifact written to
// .doug/plan/manifest.yaml after Discovery approval on a greenfield project.
type Manifest struct {
	SchemaVersion int          `yaml:"schema_version"`
	Project       Project      `yaml:"project"`
	Scaffold      Scaffold     `yaml:"scaffold"`
	Dependencies  Dependencies `yaml:"dependencies,omitempty"`
	Constraints   []string     `yaml:"constraints,omitempty"`
}

// Project holds project-level manifest fields.
type Project struct {
	Name string `yaml:"name"`
	Mode string `yaml:"mode"`
}

// Scaffold holds the technology stack fields derived from VISION.md frontmatter.
type Scaffold struct {
	Language       string `yaml:"language"`
	Runtime        string `yaml:"runtime"`
	Framework      string `yaml:"framework,omitempty"`
	PackageManager string `yaml:"package_manager,omitempty"`
	BuildSystem    string `yaml:"build_system,omitempty"`
}

// Dependencies holds the initial day-0 dependency lists.
type Dependencies struct {
	Runtime     []string `yaml:"runtime,omitempty"`
	Development []string `yaml:"development,omitempty"`
}

// ParseVisionFrontmatter extracts and parses the YAML frontmatter block from
// VISION.md content. It returns (nil, nil) when no frontmatter block is present
// — callers must treat a nil result as indicating a non-greenfield project (or
// an existing project that predates the frontmatter convention). An error is
// returned only when a frontmatter block is present but cannot be unmarshalled.
func ParseVisionFrontmatter(content string) (*VisionFrontmatter, error) {
	m := frontmatterRE.FindStringSubmatch(content)
	if m == nil {
		return nil, nil
	}
	var fm VisionFrontmatter
	if err := yaml.Unmarshal([]byte(m[1]), &fm); err != nil {
		return nil, fmt.Errorf("parsing VISION.md frontmatter: %w", err)
	}
	return &fm, nil
}

// ValidateScaffoldFrontmatter returns a human-readable error if the frontmatter
// is missing any of the required fields for greenfield manifest generation.
// Required: project_mode, language, runtime.
func ValidateScaffoldFrontmatter(fm *VisionFrontmatter) error {
	var missing []string
	if strings.TrimSpace(fm.ProjectMode) == "" {
		missing = append(missing, "project_mode")
	}
	if strings.TrimSpace(fm.Language) == "" {
		missing = append(missing, "language")
	}
	if strings.TrimSpace(fm.Runtime) == "" {
		missing = append(missing, "runtime")
	}
	if len(missing) > 0 {
		return fmt.Errorf("VISION.md frontmatter missing required fields: %s", strings.Join(missing, ", "))
	}
	return nil
}

// FromVisionFrontmatter builds a Manifest from the given project name and
// parsed VISION.md frontmatter. The caller is responsible for validating fm
// with ValidateScaffoldFrontmatter before calling this function.
func FromVisionFrontmatter(projectName string, fm *VisionFrontmatter) *Manifest {
	m := &Manifest{
		SchemaVersion: schemaVersion,
		Project: Project{
			Name: projectName,
			Mode: fm.ProjectMode,
		},
		Scaffold: Scaffold{
			Language:       fm.Language,
			Runtime:        fm.Runtime,
			Framework:      fm.Framework,
			PackageManager: fm.PackageManager,
			BuildSystem:    fm.BuildSystem,
		},
	}
	if len(fm.RuntimeDependencies) > 0 || len(fm.DevDependencies) > 0 {
		m.Dependencies = Dependencies{
			Runtime:     fm.RuntimeDependencies,
			Development: fm.DevDependencies,
		}
	}
	if len(fm.BootstrapConstraints) > 0 {
		m.Constraints = fm.BootstrapConstraints
	}
	return m
}

// Marshal serialises m to YAML bytes suitable for writing to manifest.yaml.
func Marshal(m *Manifest) ([]byte, error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshalling manifest: %w", err)
	}
	return data, nil
}

// Sync reads VISION.md from the plan directory and writes or removes
// .doug/plan/manifest.yaml depending on the project_mode in the frontmatter.
//
// For greenfield projects: validates required frontmatter fields, builds the
// manifest, and writes it atomically. Returns a human-readable error if
// required fields are missing or marshalling fails. No partial manifest is
// ever written.
//
// For non-greenfield projects (or when no frontmatter is present): removes
// .doug/plan/manifest.yaml if it exists.
func Sync(projectRoot string) error {
	visionPath := filepath.Join(layout.PlanDir(projectRoot), "VISION.md")
	data, err := os.ReadFile(visionPath)
	if err != nil {
		return fmt.Errorf("reading VISION.md: %w", err)
	}
	content := string(data)

	fm, err := ParseVisionFrontmatter(content)
	if err != nil {
		return err
	}

	manifestPath := layout.ManifestPath(projectRoot)

	// Non-greenfield or no frontmatter: remove any stale manifest.
	if fm == nil || strings.TrimSpace(fm.ProjectMode) != "greenfield" {
		if err := os.Remove(manifestPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("removing stale manifest: %w", err)
		}
		return nil
	}

	// Greenfield: validate required fields before writing anything.
	if err := ValidateScaffoldFrontmatter(fm); err != nil {
		return err
	}

	projectName := extractProjectName(content)
	m := FromVisionFrontmatter(projectName, fm)

	yamlData, err := Marshal(m)
	if err != nil {
		return err
	}

	// Atomic write: write to .tmp then rename.
	tmp := manifestPath + ".tmp"
	if err := os.WriteFile(tmp, yamlData, 0o644); err != nil {
		return fmt.Errorf("writing manifest: %w", err)
	}
	if err := os.Rename(tmp, manifestPath); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("committing manifest: %w", err)
	}
	return nil
}

// extractProjectName parses the "## Project Name" section from VISION.md
// content and returns the first non-empty line of that section, or "" if the
// section is absent or empty.
func extractProjectName(content string) string {
	// Skip past frontmatter if present.
	if m := frontmatterRE.FindStringSubmatch(content); m != nil {
		content = content[len(m[0]):]
	}

	inSection := false
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "## Project Name" {
			inSection = true
			continue
		}
		if inSection {
			if strings.HasPrefix(trimmed, "## ") {
				break
			}
			if trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}
