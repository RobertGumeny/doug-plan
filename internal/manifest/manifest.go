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
	"regexp"
	"strings"

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
