---
title: Orchestrator Loop
updated: 2026-03-18
category: Architecture
tags: [orchestrator, state, approval, agent, pipeline]
related_articles:
  - docs/kb/infrastructure/go.md
  - docs/kb/architecture/skill-system.md
---

# Orchestrator Loop

## Overview

The core orchestrator loop drives the four-stage planning pipeline. Pipeline position is always inferred from artifacts on disk — no separate state file is needed.

## Pipeline Stages

Stages are ordered. Each stage produces one artifact in `.doug/plan/`. A stage is considered done when its artifact is present.

| Stage | Artifact |
| ----- | -------- |
| Discovery | `VISION.md` |
| Roadmapping | `ROADMAP.md` |
| PRD | `PRD.md` |
| Tasks | `TASKS.md` |

`InferStage` (in `internal/state`) scans the artifacts in order and returns the first stage whose artifact is absent. If all artifacts are present, it returns `StageComplete`.

## Run Loop (`internal/orchestrator`)

Each call to `orchestrator.Run` executes one pipeline step:

1. **Re-entry**: Apply `--fresh` or `--rerun` to clear artifacts (see Re-entry Modes below).
2. **Infer stage**: `state.InferStage` reads `.doug/plan/` and returns the current stage.
3. **Write step brief**: `agent.WriteStep` creates `.doug/plan/ACTIVE_STEP.md`. It loads a stage-specific template from `internal/templates/steps/<Stage>.md` when one exists (currently `Discovery.md` and `Roadmapping.md`); otherwise a generic template is written.
4. **Invoke agent**: `agent.Invoke` runs the configured agent command as a subprocess, inheriting stdin/stdout/stderr.
5. **Parse result**: `agent.ParseResult` reads the `## Agent Result` YAML frontmatter from `ACTIVE_STEP.md` and extracts the `outcome` field.
6. **Archive step**: `agent.ArchiveStep` moves `ACTIVE_STEP.md` to `.doug/plan/logs/<stage>_<nanosecond>.md`.
7. **Dispatch outcome**:
   - `SUCCESS` → run approval gate; advance on confirmation.
   - `FAILURE` → return a non-nil error; pipeline stops.
   - `RETRY` → return nil; call `run` again to retry the same stage.

## ACTIVE_STEP.md Lifecycle

```
orchestrator.Run called
       │
       ▼
agent.WriteStep → creates .doug/plan/ACTIVE_STEP.md
       │
       ▼
agent.Invoke → agent reads and fills in ACTIVE_STEP.md
       │
       ▼
agent.ParseResult → reads outcome from ACTIVE_STEP.md
       │
       ▼
agent.ArchiveStep → moves to .doug/plan/logs/
```

The agent must write the outcome field before exiting:

```markdown
## Agent Result

---
outcome: "SUCCESS"
---
```

Valid values: `SUCCESS`, `FAILURE`, `RETRY`.

`ParseResult` searches for `## Agent Result` as a line heading (preceded by a newline). Inline references to the section name inside the Briefing text are ignored. Stage-specific step templates may mention `## Agent Result` in their briefing prose without triggering a false parse.

## Re-entry Modes

| Mode | CLI flag | Effect |
| ---- | -------- | ------ |
| Resume | (none) | No-op; `InferStage` picks up where artifacts left off |
| Re-run | `--rerun <Stage>` | Removes the named stage's artifact and all subsequent artifacts |
| Fresh | `--fresh` | Removes all pipeline artifacts; run starts at Discovery |

Stage names for `--rerun`: `Discovery`, `Roadmapping`, `PRD`, `Tasks` (case-insensitive).

## Approval Gate (`internal/approval`)

After a successful step, the gate runs before the pipeline can advance.

| Mode | Behavior |
| ---- | -------- |
| `auto` (default) | Returns immediately with no prompt |
| `soft` | Prints summary; Enter to advance, `skip` to stop |
| `hard` | Blocks until the user types `yes` |

When `soft` mode receives `skip`, `approval.Gate` returns `approval.ErrSkipped`. The orchestrator checks for this error and returns `nil` (no error, pipeline just stops).

Approval mode is resolved with CLI flag taking precedence over config:

```
CLI --approval flag  →  cfg.ApprovalMode  →  "auto"
```

## Configuration (`internal/config`)

`.doug/plan/doug-plan.yaml` keys consumed by the orchestrator:

| Key | Type | Purpose |
| --- | ---- | ------- |
| `agent` | string | Named agent: `claude`, `codex`, or `gemini` |
| `command` | []string | Full command override (takes precedence over `agent`) |
| `approval_mode` | string | Default approval mode (`auto`, `soft`, `hard`) |
| `skill_paths` | []string | Paths to skill directories (reserved for future use) |

When `command` is set, it is used verbatim. When only `agent` is set, a default command is derived:

```yaml
# Equivalent to: command: ["claude", "--print", "Please complete..."]
agent: claude
```

Unknown agents with no `command` set return an error.

## CLI Flags

```
doug-plan run [flags]

Flags:
  --approval string   approval mode override: auto, soft, or hard
  --rerun   string    re-run from stage: Discovery, Roadmapping, PRD, or Tasks
  --fresh             start fresh: clear all plan artifacts and begin at Discovery
```

## Package Map

| Package | Responsibility |
| ------- | -------------- |
| `internal/state` | Stage type, `InferStage`, `ClearArtifacts*`, `StageFromString` |
| `internal/agent` | `WriteStep`, `Invoke`, `ParseResult`, `ArchiveStep`, `Outcome` type |
| `internal/approval` | `Mode` type, `Parse`, `Gate`, `ErrSkipped` |
| `internal/config` | `Config` struct, `Load`, `AgentCommand` |
| `internal/orchestrator` | `Run`, `Options` — wires all packages together |
| `internal/templates` | Embedded `Init` FS (scaffold files) and `Steps` FS (per-stage ACTIVE_STEP.md templates) |
