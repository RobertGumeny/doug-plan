// Package handoff provides the deterministic renderer that converts per-epic
// DEFINITION.md files into PRD.md and tasks.yaml without agent invocation.
package handoff

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// EpicDefinition holds the parsed content of a per-epic DEFINITION.md.
type EpicDefinition struct {
	ID              string
	Name            string
	Overview        string
	Scope           string
	Goals           string
	NonGoals        string
	Background      string
	SuccessCriteria string
	Deliverables    string
	Tasks           []TaskDef
}

// TaskDef represents a single task parsed from the Tasks section.
type TaskDef struct {
	ID          string
	Name        string
	Type        string
	Description string
	Acceptance  []string
}

var (
	frontmatterRE  = regexp.MustCompile(`(?s)^---\n(.+?)\n---\n`)
	fmIDRE         = regexp.MustCompile(`(?m)^id:\s*"?([^"\n]+)"?\s*$`)
	fmNameRE       = regexp.MustCompile(`(?m)^name:\s*"?([^"\n]+)"?\s*$`)
	taskHeadingRE  = regexp.MustCompile(`^###\s+([\w-]+-\d+):\s*(.+)$`)
	boldFieldRE    = regexp.MustCompile(`^\*\*([^*]+)\*\*:\s*(.*)$`)
)

// Execute reads all per-epic DEFINITION.md files in planDir/epics/, renders
// per-epic PRD.md and tasks.yaml, then writes a global PRD.md and tasks.yaml
// to planDir for browser review.
func Execute(planDir string) error {
	epicsDir := filepath.Join(planDir, "epics")
	entries, err := os.ReadDir(epicsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("epics directory not found at %s", epicsDir)
		}
		return fmt.Errorf("reading epics dir: %w", err)
	}

	var defs []*EpicDefinition
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		epicID := e.Name()
		defPath := filepath.Join(epicsDir, epicID, "DEFINITION.md")
		if _, err := os.Stat(defPath); os.IsNotExist(err) {
			continue // no definition yet — skip
		}
		def, err := ParseDefinition(defPath, epicID)
		if err != nil {
			return fmt.Errorf("parsing definition for %s: %w", epicID, err)
		}
		defs = append(defs, def)

		epicDir := filepath.Join(epicsDir, epicID)
		if err := os.WriteFile(filepath.Join(epicDir, "PRD.md"), []byte(RenderPRD(def)), 0o644); err != nil {
			return fmt.Errorf("writing PRD.md for %s: %w", epicID, err)
		}
		if err := os.WriteFile(filepath.Join(epicDir, "tasks.yaml"), []byte(RenderTasksYAML(def)), 0o644); err != nil {
			return fmt.Errorf("writing tasks.yaml for %s: %w", epicID, err)
		}
	}

	if len(defs) == 0 {
		return fmt.Errorf("no epic definitions found in %s", epicsDir)
	}

	if err := os.WriteFile(filepath.Join(planDir, "PRD.md"), []byte(renderGlobalPRD(defs)), 0o644); err != nil {
		return fmt.Errorf("writing global PRD.md: %w", err)
	}
	if err := os.WriteFile(filepath.Join(planDir, "tasks.yaml"), []byte(renderGlobalTasksYAML(defs)), 0o644); err != nil {
		return fmt.Errorf("writing global tasks.yaml: %w", err)
	}
	return nil
}

// ParseDefinition parses a per-epic DEFINITION.md. epicID is used as a fallback
// when the frontmatter does not contain an id field.
func ParseDefinition(path string, epicID string) (*EpicDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	return parseDefinitionContent(string(data), epicID), nil
}

func parseDefinitionContent(content string, epicID string) *EpicDefinition {
	def := &EpicDefinition{ID: epicID, Name: epicID}

	// Extract YAML frontmatter.
	if m := frontmatterRE.FindStringSubmatch(content); m != nil {
		fm := m[1]
		if id := fmIDRE.FindStringSubmatch(fm); id != nil && strings.TrimSpace(id[1]) != "" {
			def.ID = strings.TrimSpace(id[1])
		}
		if name := fmNameRE.FindStringSubmatch(fm); name != nil && strings.TrimSpace(name[1]) != "" {
			def.Name = strings.TrimSpace(name[1])
		}
		content = content[len(m[0]):]
	}

	sections := extractSections(content)
	def.Overview = sections["Overview"]
	def.Scope = sections["Scope"]
	def.Goals = sections["Goals"]
	def.NonGoals = sections["Non-Goals"]
	def.Background = sections["Background"]
	def.SuccessCriteria = sections["Success Criteria"]
	def.Deliverables = sections["Deliverables"]
	def.Tasks = parseTasks(sections["Tasks"])

	return def
}

// extractSections splits content on ## headings and returns a map of
// heading name → trimmed section content.
func extractSections(content string) map[string]string {
	sections := make(map[string]string)
	lines := strings.Split(content, "\n")

	var current string
	var buf []string

	flush := func() {
		if current != "" {
			sections[current] = strings.TrimSpace(strings.Join(buf, "\n"))
		}
	}

	for _, line := range lines {
		if strings.HasPrefix(line, "## ") {
			flush()
			current = strings.TrimSpace(strings.TrimPrefix(line, "## "))
			buf = nil
		} else {
			buf = append(buf, line)
		}
	}
	flush()
	return sections
}

// parseTasks extracts TaskDef entries from the raw Tasks section content.
func parseTasks(content string) []TaskDef {
	if content == "" {
		return nil
	}

	var tasks []TaskDef
	var cur *TaskDef
	inAcceptance := false

	flush := func() {
		if cur != nil {
			tasks = append(tasks, *cur)
			cur = nil
		}
		inAcceptance = false
	}

	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)

		if m := taskHeadingRE.FindStringSubmatch(trimmed); m != nil {
			flush()
			cur = &TaskDef{ID: m[1], Name: strings.TrimSpace(m[2])}
			continue
		}
		if cur == nil {
			continue
		}

		if m := boldFieldRE.FindStringSubmatch(trimmed); m != nil {
			inAcceptance = false
			switch m[1] {
			case "Type":
				cur.Type = strings.TrimSpace(m[2])
			case "Description":
				cur.Description = strings.TrimSpace(m[2])
			case "Acceptance Criteria":
				inAcceptance = true
			}
			continue
		}

		if inAcceptance && strings.HasPrefix(trimmed, "- ") {
			cur.Acceptance = append(cur.Acceptance, strings.TrimPrefix(trimmed, "- "))
		}
	}
	flush()
	return tasks
}

// RenderPRD produces PRD.md content for a single epic.
func RenderPRD(def *EpicDefinition) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# PRD — %s: %s\n\n", def.ID, def.Name)
	fmt.Fprintf(&sb, "## Epic\n\n- id: %s\n- name: %q\n\n", def.ID, def.Name)

	writeSection := func(heading, content string) {
		fmt.Fprintf(&sb, "## %s\n\n", heading)
		if content != "" {
			fmt.Fprintf(&sb, "%s\n\n", content)
		}
	}

	writeSection("Overview", def.Overview)
	writeSection("Scope", def.Scope)
	writeSection("Goals", def.Goals)
	writeSection("Non-Goals", def.NonGoals)
	writeSection("Background / Context", def.Background)
	writeSection("Success Criteria", def.SuccessCriteria)
	writeSection("Deliverables", def.Deliverables)

	sb.WriteString("## Acceptance Criteria\n\n")
	for _, task := range def.Tasks {
		for _, ac := range task.Acceptance {
			fmt.Fprintf(&sb, "- %s\n", ac)
		}
	}
	sb.WriteString("\n")

	sb.WriteString("## Notes for Agents\n\nRefer to AGENTS.md for further instructions and `docs/kb` for additional context around project structure and best practices.\n")

	return sb.String()
}

// RenderTasksYAML produces tasks.yaml content for a single epic.
func RenderTasksYAML(def *EpicDefinition) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "epic:\n  id: %s\n  name: %s\n", yamlStr(def.ID), yamlStr(def.Name))

	if len(def.Tasks) == 0 {
		sb.WriteString("  tasks: []\n")
		return sb.String()
	}

	sb.WriteString("  tasks:\n")
	for _, task := range def.Tasks {
		taskType := task.Type
		if taskType == "" {
			taskType = "feature"
		}
		fmt.Fprintf(&sb, "    - id: %s\n", yamlStr(task.ID))
		fmt.Fprintf(&sb, "      type: %s\n", yamlStr(taskType))
		sb.WriteString("      status: \"TODO\"\n")
		fmt.Fprintf(&sb, "      description: %s\n", yamlStr(task.Description))
		if len(task.Acceptance) > 0 {
			sb.WriteString("      acceptance_criteria:\n")
			for _, ac := range task.Acceptance {
				fmt.Fprintf(&sb, "        - %s\n", yamlStr(ac))
			}
		} else {
			sb.WriteString("      acceptance_criteria: []\n")
		}
	}
	return sb.String()
}

// renderGlobalPRD produces a summary PRD.md listing all epics.
func renderGlobalPRD(defs []*EpicDefinition) string {
	var sb strings.Builder
	sb.WriteString("# Handoff Complete\n\n")
	sb.WriteString("All epics have been defined and handed off. Per-epic artifacts are in `.doug/plan/epics/`.\n\n")
	sb.WriteString("## Epics\n\n")
	for _, def := range defs {
		fmt.Fprintf(&sb, "### %s: %s\n\n", def.ID, def.Name)
		if def.Overview != "" {
			fmt.Fprintf(&sb, "%s\n\n", def.Overview)
		}
	}
	return sb.String()
}

// renderGlobalTasksYAML produces a consolidated tasks.yaml covering all epics.
func renderGlobalTasksYAML(defs []*EpicDefinition) string {
	var sb strings.Builder
	sb.WriteString("epics:\n")
	for _, def := range defs {
		fmt.Fprintf(&sb, "  - id: %s\n", yamlStr(def.ID))
		fmt.Fprintf(&sb, "    name: %s\n", yamlStr(def.Name))
		if len(def.Tasks) == 0 {
			sb.WriteString("    tasks: []\n")
			continue
		}
		sb.WriteString("    tasks:\n")
		for _, task := range def.Tasks {
			taskType := task.Type
			if taskType == "" {
				taskType = "feature"
			}
			fmt.Fprintf(&sb, "      - id: %s\n", yamlStr(task.ID))
			fmt.Fprintf(&sb, "        type: %s\n", yamlStr(taskType))
			sb.WriteString("        status: \"TODO\"\n")
			fmt.Fprintf(&sb, "        description: %s\n", yamlStr(task.Description))
			if len(task.Acceptance) > 0 {
				sb.WriteString("        acceptance_criteria:\n")
				for _, ac := range task.Acceptance {
					fmt.Fprintf(&sb, "          - %s\n", yamlStr(ac))
				}
			} else {
				sb.WriteString("        acceptance_criteria: []\n")
			}
		}
	}
	return sb.String()
}

// yamlStr wraps s in double quotes and escapes internal double quotes and backslashes.
func yamlStr(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}
