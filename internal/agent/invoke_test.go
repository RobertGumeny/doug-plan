package agent

import (
	"testing"
)

func TestInvoke_EmptyArgs(t *testing.T) {
	err := Invoke(t.TempDir(), []string{})
	if err == nil {
		t.Fatal("expected error for empty args, got nil")
	}
}

func TestInvoke_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	// Use `go version` as a reliable cross-platform command that exits 0.
	if err := Invoke(t.TempDir(), []string{"go", "version"}); err != nil {
		t.Fatalf("Invoke go version: %v", err)
	}
}

func TestInvoke_NonZeroExit(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	// `go` with an unknown subcommand exits non-zero.
	err := Invoke(t.TempDir(), []string{"go", "this-subcommand-does-not-exist"})
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
}
