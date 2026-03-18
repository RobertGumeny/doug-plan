package approval

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		want    Mode
		wantErr bool
	}{
		{"auto", ModeAuto, false},
		{"soft", ModeSoft, false},
		{"hard", ModeHard, false},
		{"AUTO", ModeAuto, false},
		{"SOFT", ModeSoft, false},
		{"HARD", ModeHard, false},
		{"", "", true},
		{"manual", "", true},
		{"full", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGate_Auto(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("")
	if err := Gate(ModeAuto, "Discovery", &out, in); err != nil {
		t.Fatalf("Gate(auto) returned error: %v", err)
	}
	if out.Len() != 0 {
		t.Errorf("Gate(auto) wrote output %q, want empty", out.String())
	}
}

func TestGate_Soft_Advance(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("\n")
	if err := Gate(ModeSoft, "Discovery", &out, in); err != nil {
		t.Fatalf("Gate(soft, Enter) returned error: %v", err)
	}
}

func TestGate_Soft_Skip(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("skip\n")
	err := Gate(ModeSoft, "Discovery", &out, in)
	if !errors.Is(err, ErrSkipped) {
		t.Fatalf("Gate(soft, skip) = %v, want ErrSkipped", err)
	}
}

func TestGate_Soft_SkipCaseInsensitive(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("SKIP\n")
	err := Gate(ModeSoft, "Discovery", &out, in)
	if !errors.Is(err, ErrSkipped) {
		t.Fatalf("Gate(soft, SKIP) = %v, want ErrSkipped", err)
	}
}

func TestGate_Hard_Confirm(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("yes\n")
	if err := Gate(ModeHard, "Roadmapping", &out, in); err != nil {
		t.Fatalf("Gate(hard, yes) returned error: %v", err)
	}
}

func TestGate_Hard_ConfirmCaseInsensitive(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("YES\n")
	if err := Gate(ModeHard, "Roadmapping", &out, in); err != nil {
		t.Fatalf("Gate(hard, YES) returned error: %v", err)
	}
}

func TestGate_Hard_RequiresYes(t *testing.T) {
	var out bytes.Buffer
	// First line is not "yes", second line is EOF — should error
	in := strings.NewReader("no\n")
	err := Gate(ModeHard, "Roadmapping", &out, in)
	if err == nil {
		t.Fatal("Gate(hard, no+EOF) expected error, got nil")
	}
	if errors.Is(err, ErrSkipped) {
		t.Fatal("Gate(hard) should not return ErrSkipped")
	}
}

func TestGate_Hard_RetriesUntilYes(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("no\nmaybe\nyes\n")
	if err := Gate(ModeHard, "PRD", &out, in); err != nil {
		t.Fatalf("Gate(hard, no/maybe/yes) returned error: %v", err)
	}
}

func TestGate_UnknownMode(t *testing.T) {
	var out bytes.Buffer
	in := strings.NewReader("")
	err := Gate("bogus", "Discovery", &out, in)
	if err == nil {
		t.Fatal("Gate(bogus) expected error, got nil")
	}
}
