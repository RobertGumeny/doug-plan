package agent

import (
	"fmt"
	"os"
	"os/exec"
)

// Invoke runs the agent command in the given project root directory.
// args must be a non-empty slice where args[0] is the executable name.
// The agent's stdin, stdout, and stderr are inherited from the parent process
// so interactive and headless agent sessions both work correctly.
func Invoke(projectRoot string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("agent command is empty")
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = projectRoot
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("agent %q exited with error: %w", args[0], err)
	}
	return nil
}
