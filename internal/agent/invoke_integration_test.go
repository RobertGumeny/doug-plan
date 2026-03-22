//go:build integration

package agent

import "testing"

func TestInvoke_Success(t *testing.T) {
	if err := Invoke(t.TempDir(), []string{"go", "version"}); err != nil {
		t.Fatalf("Invoke go version: %v", err)
	}
}

func TestInvoke_NonZeroExit(t *testing.T) {
	err := Invoke(t.TempDir(), []string{"go", "this-subcommand-does-not-exist"})
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
}
