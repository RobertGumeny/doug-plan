package orchestrator_test

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

// validateRoadmapFormat checks that content conforms to the hybrid
// Markdown + YAML frontmatter format specified by the Roadmapping skill:
//
//   - Top-level YAML frontmatter with project, generated, and source fields.
//   - A "# Roadmap" heading.
//   - At least three epic sections, each with an embedded YAML block
//     containing id, name, sequence, and status fields.
//   - status is always "planned".
//   - sequence values are consecutive integers starting at 1.
func validateRoadmapFormat(content string) error {
	if !strings.HasPrefix(content, "---\n") {
		return fmt.Errorf("missing opening frontmatter delimiter")
	}
	closingIdx := strings.Index(content[4:], "\n---\n")
	if closingIdx < 0 {
		return fmt.Errorf("missing closing frontmatter delimiter")
	}
	frontmatter := content[4 : 4+closingIdx]
	for _, field := range []string{"project:", "generated:", "source:"} {
		if !strings.Contains(frontmatter, field) {
			return fmt.Errorf("frontmatter missing %q field", field)
		}
	}
	if !strings.Contains(frontmatter, "source: VISION.md") {
		return fmt.Errorf("frontmatter source must be VISION.md")
	}

	if !strings.Contains(content, "# Roadmap") {
		return fmt.Errorf("missing '# Roadmap' heading")
	}

	epicHeading := regexp.MustCompile(`(?m)^## EPIC-\d+:`)
	epicCount := len(epicHeading.FindAllString(content, -1))
	if epicCount < 3 {
		return fmt.Errorf("expected at least 3 epics, found %d", epicCount)
	}

	epicBlock := regexp.MustCompile(`(?s)## (EPIC-\d+):.*?\n---\n(.*?)\n---`)
	blocks := epicBlock.FindAllStringSubmatch(content, -1)
	if len(blocks) != epicCount {
		return fmt.Errorf("epic heading/YAML-block count mismatch: %d headings, %d blocks", epicCount, len(blocks))
	}

	for i, block := range blocks {
		yaml := block[2]
		expectedSeq := strconv.Itoa(i + 1)
		for _, field := range []string{"id:", "name:", "sequence:", "status:"} {
			if !strings.Contains(yaml, field) {
				return fmt.Errorf("epic %d YAML block missing %q field", i+1, field)
			}
		}
		if !strings.Contains(yaml, "status: planned") {
			return fmt.Errorf("epic %d status must be 'planned'", i+1)
		}
		if !strings.Contains(yaml, "sequence: "+expectedSeq) {
			return fmt.Errorf("epic %d sequence must be %s", i+1, expectedSeq)
		}
	}
	return nil
}

func TestValidateRoadmapFormat_Valid(t *testing.T) {
	content := `---
project: "Test Project"
generated: "2026-03-18"
source: VISION.md
---

# Roadmap

## EPIC-1: First Epic

---
id: EPIC-1
name: "First Epic"
sequence: 1
status: planned
---

Description of first epic.

## EPIC-2: Second Epic

---
id: EPIC-2
name: "Second Epic"
sequence: 2
status: planned
---

Description of second epic.

## EPIC-3: Third Epic

---
id: EPIC-3
name: "Third Epic"
sequence: 3
status: planned
---

Description of third epic.
`
	if err := validateRoadmapFormat(content); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestValidateRoadmapFormat_Invalid(t *testing.T) {
	cases := []struct {
		name    string
		content string
	}{
		{
			name:    "missing frontmatter",
			content: "# Roadmap\n\n## EPIC-1: Foo\n",
		},
		{
			name: "missing project field",
			content: `---
generated: "2026-01-01"
source: VISION.md
---

# Roadmap

## EPIC-1: A

---
id: EPIC-1
name: "A"
sequence: 1
status: planned
---

## EPIC-2: B

---
id: EPIC-2
name: "B"
sequence: 2
status: planned
---

## EPIC-3: C

---
id: EPIC-3
name: "C"
sequence: 3
status: planned
---
`,
		},
		{
			name: "source not VISION.md",
			content: `---
project: "X"
generated: "2026-01-01"
source: OTHER.md
---

# Roadmap

## EPIC-1: A

---
id: EPIC-1
name: "A"
sequence: 1
status: planned
---

## EPIC-2: B

---
id: EPIC-2
name: "B"
sequence: 2
status: planned
---

## EPIC-3: C

---
id: EPIC-3
name: "C"
sequence: 3
status: planned
---
`,
		},
		{
			name: "fewer than 3 epics",
			content: `---
project: "X"
generated: "2026-01-01"
source: VISION.md
---

# Roadmap

## EPIC-1: A

---
id: EPIC-1
name: "A"
sequence: 1
status: planned
---
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if err := validateRoadmapFormat(tc.content); err == nil {
				t.Fatal("expected validation error, got nil")
			}
		})
	}
}
