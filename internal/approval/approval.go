package approval

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/robertgumeny/doug-plan/internal/server"
)

// Mode controls how the approval gate behaves before advancing the pipeline.
type Mode string

const (
	ModeAuto    Mode = "auto"
	ModeCLI     Mode = "cli"
	ModeBrowser Mode = "browser"
)

// ErrSkipped is returned by Gate when the user chooses to skip in cli mode.
var ErrSkipped = errors.New("step skipped by user")

// Parse parses an approval mode string. Returns an error if unrecognized.
func Parse(s string) (Mode, error) {
	switch Mode(strings.ToLower(s)) {
	case ModeAuto:
		return ModeAuto, nil
	case ModeCLI:
		return ModeCLI, nil
	case ModeBrowser:
		return ModeBrowser, nil
	default:
		return "", fmt.Errorf("unknown approval mode %q: must be auto, cli, or browser", s)
	}
}

// Gate implements the terminal approval gate for the given stage.
// Auto returns immediately. CLI prints a summary and allows a skip.
// Browser blocks until the user types "yes".
func Gate(mode Mode, stage string, out io.Writer, in io.Reader) error {
	switch mode {
	case ModeAuto:
		return nil
	case ModeCLI:
		return cliGate(stage, out, in)
	case ModeBrowser:
		return browserGate(stage, out, in)
	default:
		return fmt.Errorf("unknown approval mode %q", mode)
	}
}

// BrowserGate starts an embedded HTTP server on a dynamic port, opens the
// default browser to display the artifact for review, and blocks until the
// user POSTs approval. The approved content is written back to artifactPath.
// If secondaryPath is non-empty, that file is also served and saved on approval.
func BrowserGate(artifactPath, secondaryPath, stage string, out io.Writer) error {
	return server.Serve(artifactPath, secondaryPath, stage, out)
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func cliGate(stage string, out io.Writer, in io.Reader) error {
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

func browserGate(stage string, out io.Writer, in io.Reader) error {
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
