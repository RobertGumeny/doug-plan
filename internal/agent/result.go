package agent

import (
	"fmt"
	"os"
	"strings"

	"github.com/robertgumeny/doug-plan/internal/layout"
	"gopkg.in/yaml.v3"
)

// Outcome represents the result of a completed agent step.
type Outcome string

const (
	OutcomeSuccess Outcome = "SUCCESS"
	OutcomeFailure Outcome = "FAILURE"
	OutcomeRetry   Outcome = "RETRY"
)

// stepResult holds the parsed frontmatter from the ## Agent Result section.
type stepResult struct {
	Outcome string `yaml:"outcome"`
}

// ParseResult reads ACTIVE_STEP.md from <projectRoot>/.doug/plan/ and extracts
// the outcome value from the ## Agent Result frontmatter block.
func ParseResult(projectRoot string) (Outcome, error) {
	path := layout.ActiveStepPath(projectRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", activeStepFile, err)
	}
	return parseOutcome(string(data))
}

// parseOutcome extracts and validates the outcome from the ## Agent Result
// frontmatter block in content. Both \n and \r\n line endings are handled.
func parseOutcome(content string) (Outcome, error) {
	// Search for "## Agent Result" as a line heading (preceded by a newline)
	// so that inline references to it in Briefing text are not mistaken for
	// the actual section header.
	const sectionHeading = "\n## Agent Result"
	idx := strings.Index(content, sectionHeading)
	if idx == -1 {
		return "", fmt.Errorf("ACTIVE_STEP.md: missing %q section", "## Agent Result")
	}
	rest := content[idx+len(sectionHeading):]

	// Find the opening --- delimiter after the section header.
	first := strings.Index(rest, "---")
	if first == -1 {
		return "", fmt.Errorf("ACTIVE_STEP.md: missing opening --- in Agent Result section")
	}
	rest = rest[first+3:]

	// Find the closing --- delimiter.
	second := strings.Index(rest, "---")
	if second == -1 {
		return "", fmt.Errorf("ACTIVE_STEP.md: missing closing --- in Agent Result section")
	}
	frontmatter := strings.TrimSpace(rest[:second])

	var result stepResult
	if err := yaml.Unmarshal([]byte(frontmatter), &result); err != nil {
		return "", fmt.Errorf("parsing Agent Result frontmatter: %w", err)
	}

	outcome := Outcome(strings.TrimSpace(result.Outcome))
	switch outcome {
	case OutcomeSuccess, OutcomeFailure, OutcomeRetry:
		return outcome, nil
	case "":
		return "", fmt.Errorf("ACTIVE_STEP.md: outcome is empty")
	default:
		return "", fmt.Errorf("ACTIVE_STEP.md: unknown outcome %q", outcome)
	}
}
