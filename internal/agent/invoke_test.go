package agent

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"
)

func TestInvoke_EmptyArgs(t *testing.T) {
	err := Invoke(t.TempDir(), []string{})
	if err == nil {
		t.Fatal("expected error for empty args, got nil")
	}
}

func TestBuildCommand_WiresProcessAttributes(t *testing.T) {
	projectRoot := t.TempDir()
	var stdin bytes.Buffer
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	cmd, err := buildCommand(projectRoot, []string{"go", "version"}, &stdin, &stdout, &stderr)
	if err != nil {
		t.Fatalf("buildCommand: %v", err)
	}
	if len(cmd.Args) == 0 || cmd.Args[0] != "go" {
		t.Fatalf("cmd.Args[0] = %q, want %q", cmd.Args[0], "go")
	}
	if len(cmd.Args) != 2 || cmd.Args[1] != "version" {
		t.Fatalf("cmd.Args = %v, want [go version]", cmd.Args)
	}
	if cmd.Dir != projectRoot {
		t.Fatalf("cmd.Dir = %q, want %q", cmd.Dir, projectRoot)
	}
	if cmd.Stdin != &stdin {
		t.Fatal("stdin was not wired through to exec.Cmd")
	}
	if cmd.Stdout != &stdout {
		t.Fatal("stdout was not wired through to exec.Cmd")
	}
	if cmd.Stderr != &stderr {
		t.Fatal("stderr was not wired through to exec.Cmd")
	}
}

func TestInvoke_WrapsRunnerError(t *testing.T) {
	oldRunCmd := runCmd
	runCmd = func(cmd *exec.Cmd) error {
		if cmd.Dir == "" {
			t.Fatal("expected working directory to be set before run")
		}
		return errors.New("boom")
	}
	defer func() {
		runCmd = oldRunCmd
	}()

	err := Invoke(t.TempDir(), []string{"go", "version"})
	if err == nil {
		t.Fatal("expected runner error, got nil")
	}
	if got := err.Error(); got != `agent "go" exited with error: boom` {
		t.Fatalf("Invoke error = %q, want wrapped runner error", got)
	}
}
