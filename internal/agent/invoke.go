package agent

import (
	"fmt"
	"io"
	"os"
	"os/exec"
)

var runCmd = func(cmd *exec.Cmd) error {
	return cmd.Run()
}

// Invoke runs the agent command in the given project root directory.
// args must be a non-empty slice where args[0] is the executable name.
// The agent's stdin, stdout, and stderr are inherited from the parent process
// so interactive and headless agent sessions both work correctly.
func Invoke(projectRoot string, args []string) error {
	cmd, err := buildCommand(projectRoot, args, os.Stdin, os.Stdout, os.Stderr)
	if err != nil {
		return err
	}
	if err := runCmd(cmd); err != nil {
		return fmt.Errorf("agent %q exited with error: %w", args[0], err)
	}
	return nil
}

func buildCommand(projectRoot string, args []string, stdin io.Reader, stdout, stderr io.Writer) (*exec.Cmd, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("agent command is empty")
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = projectRoot
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd, nil
}
