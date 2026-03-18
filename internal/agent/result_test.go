package agent

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseOutcome(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Outcome
		wantErr bool
	}{
		{
			name: "SUCCESS",
			content: `# Active Step
## Agent Result
---
outcome: "SUCCESS"
---
## Output
`,
			want: OutcomeSuccess,
		},
		{
			name: "FAILURE",
			content: `# Active Step
## Agent Result
---
outcome: "FAILURE"
---
`,
			want: OutcomeFailure,
		},
		{
			name: "RETRY",
			content: `# Active Step
## Agent Result
---
outcome: "RETRY"
---
`,
			want: OutcomeRetry,
		},
		{
			name: "empty outcome",
			content: `# Active Step
## Agent Result
---
outcome: ""
---
`,
			wantErr: true,
		},
		{
			name: "unknown outcome",
			content: `# Active Step
## Agent Result
---
outcome: "DONE"
---
`,
			wantErr: true,
		},
		{
			name:    "missing Agent Result section",
			content: `# Active Step\n## Output\n`,
			wantErr: true,
		},
		{
			name: "missing opening delimiter",
			content: `# Active Step
## Agent Result
outcome: "SUCCESS"
---
`,
			wantErr: true,
		},
		{
			name: "missing closing delimiter",
			content: `# Active Step
## Agent Result
---
outcome: "SUCCESS"
`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseOutcome(tt.content)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseOutcome() expected error, got nil (outcome=%q)", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseOutcome() unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("parseOutcome() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseResult(t *testing.T) {
	root := t.TempDir()
	dougDir := filepath.Join(root, ".doug")
	if err := os.MkdirAll(dougDir, 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}

	content := `# Active Step
## Agent Result
---
outcome: "SUCCESS"
---
## Output
`
	if err := os.WriteFile(filepath.Join(dougDir, activeStepFile), []byte(content), 0o644); err != nil {
		t.Fatalf("setup: writing ACTIVE_STEP.md: %v", err)
	}

	got, err := ParseResult(root)
	if err != nil {
		t.Fatalf("ParseResult: %v", err)
	}
	if got != OutcomeSuccess {
		t.Errorf("ParseResult() = %q, want %q", got, OutcomeSuccess)
	}
}

func TestParseResult_MissingFile(t *testing.T) {
	root := t.TempDir()
	_, err := ParseResult(root)
	if err == nil {
		t.Fatal("expected error for missing ACTIVE_STEP.md, got nil")
	}
}
