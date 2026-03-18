---
title: Go Infrastructure & Best Practices
updated: 2026-03-17
category: Infrastructure
tags: [go, golang, build, testing]
related_articles:
  - docs/kb/dependencies/go-1-26.md
---

# Go Infrastructure & Best Practices

## Overview

doug-plan is built with Go 1.26, the current stable release as of project start. All contributors should be on 1.26 or newer.

```bash
go version   # should output go1.26.x or higher
```

The `go.mod` minimum version is pinned to `1.26`. Do not lower it.

## Module Path

```
github.com/robertgumeny/doug-plan
```

Replace `robertgumeny` if forked. All internal imports use this path.

## Project Structure

```
doug-plan/
├── cmd/
│   ├── root.go         # rootCmd definition and Execute()
│   ├── init.go         # init subcommand — project scaffolding
│   └── run.go          # run subcommand — --approval, --rerun, --fresh flags
├── internal/
│   ├── layout/         # Shared path helpers for .doug/plan-owned files
│   ├── scaffold/       # scaffold.Run() — creates .doug/plan/, AGENTS.md, CLAUDE.md, and agent skill dirs
│   ├── config/         # Config struct, Load, AgentCommand — reads .doug/plan/doug-plan.yaml
│   ├── orchestrator/   # Run(Options) — full pipeline loop (EPIC-2)
│   ├── agent/          # WriteStep, Invoke, ParseResult, ArchiveStep, Outcome type
│   ├── approval/       # Gate (auto/soft/hard), Parse, ErrSkipped
│   ├── state/          # Stage type, InferStage, ClearArtifacts*, StageFromString
│   └── templates/      # Embedded init templates for AGENTS/CLAUDE/provider scaffolding
├── main.go             # One line: cmd.Execute()
```

**Rule**: `cmd/` wires things together. All logic lives in `internal/`. If a function in `cmd/` is doing more than calling into `internal/`, it belongs in a package.

## Dependencies

Current approved dependencies:

| Package                  | Purpose                                   |
| ------------------------ | ----------------------------------------- |
| `github.com/spf13/cobra` | CLI framework (`run`, `init` subcommands) |
| `gopkg.in/yaml.v3`       | YAML marshal/unmarshal for state files    |

Everything else should be stdlib. In particular:

**No go-git** — all git operations use `exec.Command("git", ...)` with an explicit args slice.

**No logging library** — use a custom `internal/log` package with ANSI codes and stdlib only when logging is needed.

**No alternative YAML libraries** — do not introduce `goccy/go-yaml` or `sigs.k8s.io/yaml`.

When adding a new dependency, run `go mod tidy` before writing your session result. The orchestrator's install step runs `go mod download`, which only downloads modules already listed in `go.mod` — it does not resolve new imports from source. You must run `go mod tidy` yourself.

## Key Decisions

**`exec.Command` over shell eval**: Never use `sh -c` or string concatenation to build shell commands. Always pass an explicit args slice. This is a hard rule — it applies to git, build commands, and agent invocation.

**Atomic file writes**: All state file writes go to a `.tmp` file first, then `os.Rename` to the final path. This prevents partial writes from corrupting state files if the process is killed mid-write.

**Single `SaveState()` call per iteration**: Load state structs once, mutate in memory, write once. Never multiple sequential mutations to the same file.

**Three failure tiers**: Unambiguous self-correction is silent (Tier 1), recoverable-with-risk emits a warning (Tier 2), ambiguous or git-state-touching failures exit loudly with a clear message (Tier 3). Before any self-correction, ask: could this same condition re-trigger next iteration? If yes, Tier 3.

## Implementation

**Exec commands:**

```go
// Good
cmd := exec.Command("git", "commit", "-m", message)
cmd.Dir = projectRoot

// Bad — shell injection risk, not cross-platform
cmd := exec.Command("sh", "-c", "git commit -m "+message)
```

**Atomic file write:**

```go
tmp := path + ".tmp"
if err := os.WriteFile(tmp, data, 0644); err != nil {
    return err
}
return os.Rename(tmp, path)
```

**Error wrapping:**

```go
// Good — enough context for the caller to log without re-wrapping
return fmt.Errorf("loading project state from %s: %w", path, err)

// Too vague
return fmt.Errorf("failed to load file: %w", err)
```

**Best-effort terminal or injected-writer output:**

```go
func writef(w io.Writer, format string, args ...any) {
    _, _ = fmt.Fprintf(w, format, args...)
}

// Good: prompt or summary output that should not change command flow
writef(os.Stdout, "Selection (1-4, or press Enter for go): ")
```

Use this pattern only for informational output where a write error is intentionally non-fatal. Persisted file writes and other correctness-affecting I/O must still return errors.

**Failure tier mapping:**

```go
// Tier 1: handle internally, return nothing
func fixAttemptCounter(state *types.ProjectState) {
    state.ActiveTask.Attempts--
}

// Tier 2: return a warning result, not an error
type ValidationResult struct { AutoCorrected bool; Description string }

// Tier 3: return a non-nil error; main loop calls log.Fatal
return fmt.Errorf("nested bug detected during bugfix task %s — manual intervention required", taskID)
```

**Table-driven tests:**

```go
tests := []struct {
    name    string
    input   string
    want    string
    wantErr bool
}{
    {"valid SUCCESS", "SUCCESS", "SUCCESS", false},
    {"empty outcome", "", "", true},
    {"unknown outcome", "DONE", "", true},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) { ... })
}
```

**Integration test skip:**

```go
if testing.Short() {
    t.Skip("skipping integration test")
}
```

## Go 1.26 Features Relevant to doug-plan

**Green Tea GC (now default)**: Reduces GC overhead by 10–40% for allocation-heavy programs. doug-plan's YAML struct allocations and file I/O benefit from this automatically. To disable if you see a regression: `GOEXPERIMENT=nogreenteagc` at build time.

**`new()` accepts expressions**: Useful for optional pointer fields in structs. `new(someExpression)` allocates a pointer to the result. Use it where it reduces boilerplate on state optional fields.

**`go fix` is now a modernizer**: Rewritten on the same analysis framework as `go vet`. Run `go fix ./...` periodically — fixers are behavior-preserving and update idioms automatically.

**Stack-allocated slice backing stores**: The compiler stack-allocates slice backing stores in more cases. Short-lived slices in the hot loop are cheaper with no code changes needed.

**Faster small allocations**: Size-specialized malloc reduces allocations under 512 bytes by up to 30%.

## Build

| Command      | Effect                              |
| ------------ | ----------------------------------- |
| `make build` | `go build -o doug-plan .`           |
| `make test`  | `go test ./...`                     |
| `make lint`  | `go vet ./...`                      |
| `make clean` | `rm -f doug-plan`                   |

## Edge Cases & Gotchas

**`go.sum` presence**: A project with `go.mod` but no `go.sum` has not had `go mod tidy` run. Ensure `go.sum` is committed before starting tasks that depend on installed dependencies.

**Cross-platform paths**: Use `filepath.Join` everywhere — never string concatenation. Use `os.Executable()` or pass `projectRoot` explicitly as a parameter. Never use `os.Getwd()` as a proxy for project root; it breaks when the binary is invoked from a different directory.

**Line endings**: When parsing agent-written files, handle both `\r\n` and `\n`. Agents running on Windows will produce CRLF.

**`go mod download` vs `go mod tidy`**: `go mod download` only fetches modules already in `go.mod`. If you added a new import in source code, you must run `go mod tidy` yourself first, or the subsequent build will fail.

## Useful Commands

```bash
# Modernize code to current idioms
go fix ./...

# Check for issues
go vet ./...

# Tidy after adding a new import
go mod tidy

# Build for a specific platform
GOOS=windows GOARCH=amd64 go build -o doug-plan.exe .

# Run only unit tests (skip integration)
go test -short ./...

# Run everything including integration
go test ./...
```

## Related Topics

- [Go 1.26 Dependency](../dependencies/go-1-26.md) — version pinning and upgrade notes
