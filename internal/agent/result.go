package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

// ParseResult reads ACTIVE_STEP.md from <projectRoot>/.doug/ and extracts
// the outcome value from the ## Agent Result frontmatter block.
func ParseResult(projectRoot string) (Outcome, error) {
	path := filepath.Join(projectRoot, ".doug", activeStepFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("reading %s: %w", activeStepFile, err)
	}
	return parseOutcome(string(data))
}

// parseOutcome extracts and validates the outcome from the ## Agent Result
// frontmatter block in content. Both \n and \r\n line endings are handled.
func parseOutcome(content string) (Outcome, error) {
	const section = "## Agent Result"
	idx := strings.Index(content, section)
	if idx == -1 {
		return "", fmt.Errorf("ACTIVE_STEP.md: missing %q section", section)
	}
	rest := content[idx+len(section):]

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
