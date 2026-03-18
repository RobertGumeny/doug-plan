package approval

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
)

// Mode controls how the approval gate behaves before advancing the pipeline.
type Mode string

const (
	ModeAuto Mode = "auto"
	ModeSoft Mode = "soft"
	ModeHard Mode = "hard"
)

// ErrSkipped is returned by Gate when the user chooses to skip in soft mode.
var ErrSkipped = errors.New("step skipped by user")

// Parse parses an approval mode string. Returns an error if unrecognized.
func Parse(s string) (Mode, error) {
	switch Mode(strings.ToLower(s)) {
	case ModeAuto:
		return ModeAuto, nil
	case ModeSoft:
		return ModeSoft, nil
	case ModeHard:
		return ModeHard, nil
	default:
		return "", fmt.Errorf("unknown approval mode %q: must be auto, soft, or hard", s)
	}
}

// Gate implements the terminal approval gate for the given stage.
// Auto returns immediately. Soft prints a summary and allows a skip.
// Hard blocks until the user types "yes".
func Gate(mode Mode, stage string, out io.Writer, in io.Reader) error {
	switch mode {
	case ModeAuto:
		return nil
	case ModeSoft:
		return softGate(stage, out, in)
	case ModeHard:
		return hardGate(stage, out, in)
	default:
		return fmt.Errorf("unknown approval mode %q", mode)
	}
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func softGate(stage string, out io.Writer, in io.Reader) error {
	writef(out, "\nStep %s completed. Press Enter to advance, or type 'skip' to stop here: ", stage)
	scanner := bufio.NewScanner(in)
	if scanner.Scan() {
		if strings.ToLower(strings.TrimSpace(scanner.Text())) == "skip" {
			writef(out, "Stopping after %s. Run again to advance to the next step.\n", stage)
			return ErrSkipped
		}
	}
	return nil
}

func hardGate(stage string, out io.Writer, in io.Reader) error {
	writef(out, "\nStep %s completed. Type 'yes' to advance the pipeline: ", stage)
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		if strings.ToLower(strings.TrimSpace(scanner.Text())) == "yes" {
			return nil
		}
		writef(out, "Type 'yes' to advance the pipeline: ")
	}
	return fmt.Errorf("approval gate: no confirmation received for stage %s", stage)
}
