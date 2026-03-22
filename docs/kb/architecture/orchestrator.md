---
title: Orchestrator Loop
updated: 2026-03-22
category: Architecture
tags: [orchestrator, state, approval, agent, pipeline]
related_articles:
  - docs/kb/infrastructure/go.md
  - docs/kb/architecture/skill-system.md
  - docs/kb/architecture/browser-ui.md
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
| Definition | `DEFINITION.md` |
| PRD | `PRD.md` |
| Tasks | `tasks.yaml` |

`InferStage` (in `internal/state`) scans the artifacts in order and returns the first stage whose artifact is absent **or invalid**. If all artifacts are present and valid, it returns `StageComplete`. Each artifact has a content validator (see `internal/state/validate.go`); an unfilled shell is treated the same as a missing file so the pipeline re-enters the correct stage.

## Run Loop (`internal/orchestrator`)

Each call to `orchestrator.Run` executes one pipeline step:

1. **Re-entry**: Apply `--fresh` or `--rerun` to clear artifacts (see Re-entry Modes below).
2. **Infer stage**: `state.InferStage` reads `.doug/plan/` and returns the current stage.
3. **StagePRD shortcut**: If the current stage is `PRD`, the orchestrator calls `handoff.Execute` (the deterministic renderer) directly and then runs the approval gate. No agent is invoked; steps 4–8 are skipped.
4. **Materialize artifact shell**: `agent.MaterializeArtifact` writes the host-owned artifact template to `.doug/plan/<artifact>` if the file does not already exist. Stages without an embedded shell template (PRD, Tasks) are a no-op. This gives the agent a pre-defined document structure to fill rather than requiring it to invent the shape.
5. **Write step brief**: `agent.WriteStep` creates `.doug/plan/ACTIVE_STEP.md`. It loads a stage-specific template from `internal/templates/steps/<Stage>.md` when one exists (currently `Discovery.md`, `Roadmapping.md`, and `Definition.md`); otherwise a generic template is written.
6. **Invoke agent**: `agent.Invoke` runs the configured agent command as a subprocess, inheriting stdin/stdout/stderr.
7. **Parse result**: `agent.ParseResult` reads the `## Agent Result` YAML frontmatter from `ACTIVE_STEP.md` and extracts the `outcome` field.
8. **Archive step**: `agent.ArchiveStep` moves `ACTIVE_STEP.md` to `.doug/plan/logs/<stage>_<nanosecond>.md`.
9. **Dispatch outcome**:
   - `SUCCESS` → run approval gate; advance on confirmation.
   - `FAILURE` → return a non-nil error; pipeline stops.
   - `RETRY` → remove the artifact shell (so `InferStage` re-enters this stage on the next run) and return nil.

## ACTIVE_STEP.md Lifecycle

ACTIVE_STEP.md is only written for agent-driven stages (Discovery, Roadmapping, Definition). StagePRD uses the deterministic handoff renderer and does not write ACTIVE_STEP.md.

```
orchestrator.Run called
       │
       ▼
agent.MaterializeArtifact → writes artifact shell if absent
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

Stage names for `--rerun`: `Discovery`, `Roadmapping`, `Definition`, `PRD`, `Tasks` (case-insensitive).

## Approval Gate (`internal/approval`)

After a successful step, the gate runs before the pipeline can advance.

| Mode | Behavior |
| ---- | -------- |
| `auto` (default) | Returns immediately with no prompt |
| `soft` | Prints summary; Enter to advance, `skip` to stop |
| `hard` | Opens browser review UI; blocks until user clicks Approve |

In `hard` mode the orchestrator calls `approval.BrowserGate`, which starts an embedded HTTP server, opens the browser, and blocks until `POST /approve` is received. The approved content is written back to disk before the pipeline advances. See [Browser UI](browser-ui.md) for the full flow.

In `soft` mode, when the user types `skip`, `approval.Gate` returns `approval.ErrSkipped`. The orchestrator checks for this error and returns `nil` (no error, pipeline just stops).

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
  --rerun   string    re-run from stage: Discovery, Roadmapping, Definition, PRD, or Tasks
  --fresh             start fresh: clear all plan artifacts and begin at Discovery
```

## Package Map

| Package | Responsibility |
| ------- | -------------- |
| `internal/state` | Stage type, `InferStage`, `ArtifactFile`, `ClearArtifacts*`, `StageFromString` |
| `internal/agent` | `WriteStep`, `MaterializeArtifact`, `Invoke`, `ParseResult`, `ArchiveStep`, `Outcome` type |
| `internal/handoff` | `Execute` — deterministic renderer: parses per-epic `DEFINITION.md` files and writes `PRD.md` and `tasks.yaml` |
| `internal/approval` | `Mode` type, `Parse`, `Gate`, `BrowserGate`, `ErrSkipped` |
| `internal/config` | `Config` struct, `Load`, `AgentCommand` |
| `internal/orchestrator` | `Run`, `Options` — wires all packages together |
| `internal/server` | Embedded HTTP server for browser review (`Serve`) |
| `internal/ui` | `Bundle embed.FS` — compiled React bundle (`bundle.html`) |
| `internal/templates` | Embedded `Init` FS (scaffold files), `Steps` FS (per-stage ACTIVE_STEP.md templates), and `Artifacts` FS (artifact shell templates) |
