package state

import (
	"fmt"
	"strings"
)

// validators maps artifact filenames to their content validation functions.
// A validator returns a non-nil error when the artifact content is invalid or
// incomplete (e.g. still the unfilled shell written by MaterializeArtifact).
// InferStage calls the appropriate validator after confirming the file exists;
// an invalid artifact is treated the same as a missing one.
var validators = map[string]func(string) error{
	"VISION.md":     validateVision,
	"ROADMAP.md":    validateRoadmap,
	"DEFINITION.md": validateDefinition,
	"PRD.md":        validatePRD,
	"tasks.yaml":    validateTasksYAML,
}

// validateVision returns an error if VISION.md appears to be an unfilled shell.
// A valid VISION.md must contain at least one non-heading, non-blank line.
// The artifact shell written by MaterializeArtifact contains only headings and
// blank lines, so any substantive agent-written content satisfies this check.
func validateVision(content string) error {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			return nil
		}
	}
	return fmt.Errorf("VISION.md: no content beyond headings — may be an unfilled shell")
}

// validateRoadmap returns an error if ROADMAP.md has an empty project field.
// The artifact shell written by MaterializeArtifact always contains project: "".
func validateRoadmap(content string) error {
	if strings.Contains(content, `project: ""`) {
		return fmt.Errorf("ROADMAP.md: frontmatter project field is empty")
	}
	return nil
}

// validateDefinition returns an error if DEFINITION.md has an empty id field.
// The artifact shell written by MaterializeArtifact always contains id: "".
func validateDefinition(content string) error {
	if strings.Contains(content, `id: ""`) {
		return fmt.Errorf("DEFINITION.md: frontmatter id field is empty")
	}
	return nil
}

// validatePRD returns an error if PRD.md does not start with a top-level heading.
// PRD.md is produced by the deterministic handoff renderer, which always writes
// a document beginning with "# ".
func validatePRD(content string) error {
	if !strings.HasPrefix(content, "# ") {
		return fmt.Errorf("PRD.md: missing top-level heading")
	}
	return nil
}

// validateTasksYAML returns an error if tasks.yaml does not begin with the
// "epics:" root key. tasks.yaml is produced by the deterministic handoff
// renderer, which always writes a document starting with "epics:".
func validateTasksYAML(content string) error {
	if !strings.HasPrefix(strings.TrimSpace(content), "epics:") {
		return fmt.Errorf("tasks.yaml: missing 'epics:' root key")
	}
	return nil
}
