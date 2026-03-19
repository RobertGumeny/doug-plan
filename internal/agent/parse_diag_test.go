package agent

import "testing"

// TestParseOutcome_InlineReference verifies that parseOutcome finds the
// ## Agent Result section heading even when the briefing text contains an
// inline reference to "## Agent Result" (e.g. "write the outcome into the
// ## Agent Result block"). This was a latent bug found during EPIC-3-005
// pipeline validation.
func TestParseOutcome_InlineReference(t *testing.T) {
	// Mirrors the real Discovery ACTIVE_STEP.md format: the Briefing section
	// mentions "## Agent Result" inline, then the actual section follows.
	content := "# Active Step\n\n## Briefing\n\n4. Write the outcome into this file's `## Agent Result` block before exiting.\n\n---\n\n## Agent Result\n\n---\noutcome: \"SUCCESS\"\n---\n\n## Output\n"

	got, err := parseOutcome(content)
	if err != nil {
		t.Fatalf("parseOutcome: %v", err)
	}
	if got != OutcomeSuccess {
		t.Errorf("outcome = %q, want %q", got, OutcomeSuccess)
	}
}
