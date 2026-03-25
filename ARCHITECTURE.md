# Architecture

This document describes the `doug-plan` system for first-time contributors. It covers the pipeline, major packages, key data flows, and design decisions.

---

## High-Level Overview

`doug-plan` is a CLI binary written in Go. It runs a five-stage planning pipeline where the first four stages are agent-driven and the fifth (Handoff) is deterministic. Each stage produces one artifact file on disk; pipeline position is always inferred from which artifacts are present and valid.

```
┌─────────────────────────────────────────────────────────────────┐
│  CLI (cmd/)                                                      │
│    init → scaffold.Run                                           │
│    run  → orchestrator.Run (executes one step, then exits)       │
└────────────────────────┬────────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────────┐
│  Orchestrator (internal/orchestrator)                            │
│                                                                  │
│  1. state.InferStage   → which stage needs to run?              │
│  2. agent.WriteStep    → write ACTIVE_STEP.md (brief for agent) │
│  3. agent.Invoke       → run agent subprocess                    │
│  4. agent.ParseResult  → read outcome from ACTIVE_STEP.md       │
│  5. agent.ArchiveStep  → move ACTIVE_STEP.md to logs/           │
│  6. approval.Gate      → human review (auto/cli/browser)        │
└─────────────────────────────────────────────────────────────────┘
```

---

## Pipeline Stages

Stages run in order. Each stage is done when its artifact is present **and valid**.

| Stage | Artifact | Driver |
| ----- | -------- | ------ |
| Discovery | `.doug/plan/VISION.md` | Agent |
| Roadmapping | `.doug/plan/ROADMAP.md` | Agent |
| Definition | `.doug/plan/DEFINITION.md` | Agent |
| PRD | `.doug/plan/PRD.md` | Agent (formerly); now deterministic |
| Tasks | `.doug/plan/tasks.yaml` | Deterministic renderer |

`state.InferStage` scans these artifacts in order and returns the first whose file is missing or fails its content validator. An unfilled shell (created by the host before agent invocation) counts as invalid, so the pipeline always enters the correct stage even if shells were pre-created.

---

## Package Map

```
doug-plan/
├── main.go                    # cmd.Execute() — nothing else
├── cmd/
│   ├── root.go                # rootCmd, Execute()
│   ├── init.go                # init subcommand → scaffold.Run
│   └── run.go                 # run subcommand → orchestrator.Run; --approval, --rerun, --fresh flags
└── internal/
    ├── orchestrator/          # Run(Options) — executes one pipeline step per call
    ├── state/                 # Stage type, InferStage, ArtifactFile, ClearArtifacts*, StageFromString, validators
    ├── agent/                 # WriteStep, Invoke, ParseResult, ArchiveStep, Outcome type
    ├── approval/              # Mode type, Parse, Gate (auto/cli/browser), BrowserGate, ErrSkipped
    ├── config/                # Config struct, Load, AgentCommand — reads doug-plan.yaml
    ├── prompt/                # SelectOne, Text, IsTTY — reusable interactive prompt helpers
    ├── scaffold/              # scaffold.Run — creates .doug/plan/, AGENTS.md, skill dirs
    ├── server/                # Embedded HTTP server for browser review (Serve)
    ├── ui/                    # Bundle embed.FS — compiled React bundle (bundle.html)
    ├── layout/                # Shared path helpers for .doug/plan-owned files
    └── templates/             # Embedded FS: init scaffold files and per-stage ACTIVE_STEP.md templates
        ├── init/              # Files copied by scaffold.Run (AGENTS.md, CLAUDE.md, skills/)
        ├── steps/             # Per-stage ACTIVE_STEP.md templates (Discovery.md, Roadmapping.md, …)
        └── artifacts/        # Per-stage artifact shells (VISION.md, ROADMAP.md, DEFINITION.md)
```

**Rule**: `cmd/` only wires things together. All logic lives in `internal/`. If a function in `cmd/` is doing more than delegating to `internal/`, it belongs in a package.

---

## Orchestrator Loop

`orchestrator.Run` drives one pipeline step per call. `doug-plan run` invokes it once, returns, and a later `doug-plan run` continues from the next inferred stage:

```
Run called
  │
  ├─ apply --fresh or --rerun flags (clear artifacts)
  ├─ state.InferStage         → which stage?
  ├─ agent.WriteStep          → write .doug/plan/ACTIVE_STEP.md
  ├─ agent.Invoke             → exec agent subprocess (inherits stdin/stdout/stderr)
  ├─ agent.ParseResult        → read outcome field from ACTIVE_STEP.md
  ├─ agent.ArchiveStep        → move ACTIVE_STEP.md to .doug/plan/logs/<stage>_<ns>.md
  └─ dispatch outcome
       SUCCESS  → run approval gate; advance on confirmation
       FAILURE  → return error; pipeline stops
       RETRY    → return nil; same stage runs again on next call
```

### ACTIVE_STEP.md

`ACTIVE_STEP.md` is the brief the agent reads and writes. `WriteStep` loads a stage-specific template from `internal/templates/steps/<Stage>.md`; if no template exists for that stage a generic one is written.

The agent must write an `## Agent Result` block before exiting:

```markdown
## Agent Result

---
outcome: "SUCCESS"
---
```

`ParseResult` searches for `## Agent Result` as a **line heading** (preceded by a newline) to avoid false matches when the section name appears in briefing prose.

### Re-entry Modes

| Mode | Flag | Effect |
| ---- | ---- | ------ |
| Resume | (none) | `InferStage` picks up where artifacts left off |
| Re-run | `--rerun <Stage>` | Removes that stage's artifact and all subsequent ones |
| Fresh | `--fresh` | Removes all pipeline artifacts; restarts at Discovery |

---

## Configuration

`.doug/plan/doug-plan.yaml` controls agent selection and approval mode:

```yaml
agent: claude          # claude, codex, or gemini
approval_mode: auto    # auto, cli, or browser
# command: [...]       # full command override; takes precedence over agent
```

`config.AgentCommand` derives the subprocess command from `agent` if no `command` is set:

```
agent: claude  →  ["claude", "Please complete..."]
```

An unknown agent name with no `command` override returns an error.

---

## Approval Gate

After each `SUCCESS` outcome the orchestrator runs an approval gate before advancing.

| Mode | Behavior |
| ---- | -------- |
| `auto` | Returns immediately; no interaction |
| `cli` | Prints summary; Enter to advance, `skip` to stop (returns `ErrSkipped`) |
| `browser` | Opens browser review UI; blocks until user clicks Approve |

In `browser` mode `approval.BrowserGate` delegates to `server.Serve`:

```
server.Serve starts HTTP listener on 127.0.0.1:<random-port>
  │
  ├─ prints "Review URL: http://127.0.0.1:<port>"
  ├─ opens browser (best-effort; failures are silent)
  │
  ├─ GET /          → serves bundle.html (embedded React app)
  ├─ GET /artifact  → returns {stage, content, secondaryContent} as JSON
  │
  └─ POST /approve  → writes content to disk (atomic tmp→rename); shuts down server
```

The server shuts down after the first `POST /approve`. The URL is always printed as a fallback.

---

## Browser UI

The React app is a single self-contained HTML file embedded in the binary via `embed.FS`. It has no runtime external dependencies and requires no network access.

**Per-stage views:**

| Stage | View | Description |
| ----- | ---- | ----------- |
| Discovery | `VisionView` | Large textarea for VISION.md |
| Roadmapping | `RoadmapView` | Epic card list with editable title, description, reorder |
| Definition | `DefinitionView` | Guided per-task fields |
| PRD | `PRDView` | Split prose textarea and task list |
| Tasks | `TasksView` | Structured task review |

**Build pipeline:**

```
make build-ui
  npm install --prefix ui
  node ui/build.js          ← esbuild bundles ui/src/index.jsx as IIFE
                             → writes internal/ui/bundle.html

make build
  (runs build-ui)
  go build -o doug-plan .
```

`bundle.html` is committed to the repository. `go build` alone works without Node.js as long as the bundle is up to date.

---

## Skill System

Skills are Markdown files that give an AI agent a structured workflow. During `doug init`, `scaffold.Run` copies skill templates from `internal/templates/init/skills/` into each agent's skill directory.

**Skill file format:**

```markdown
---
name: "skill-name"
description: "One-line description."
---

# Skill Title

Phases and steps the agent follows.
```

**Agent directories after `doug init --agents claude,codex`:**

```
.claude/skills/discovery/SKILL.md
.claude/skills/roadmapping/SKILL.md
.claude/skills/definition/SKILL.md
.claude/skills/handoff/SKILL.md
.claude/skills/research/SKILL.md
.codex/skills/...  (same set)
```

Agents invoke skills via slash commands (e.g. `/discovery`, `/research`). Skills are agent-local copies — changes to templates in the source tree do not retroactively update existing initialized projects.

**Planning skills:**

| Skill | Reads | Writes |
| ----- | ----- | ------ |
| `discovery` | (user interview) | `VISION.md` |
| `roadmapping` | `VISION.md` | `ROADMAP.md` |
| `definition` | `VISION.md`, `ROADMAP.md` | `DEFINITION.md` |
| `handoff` | `DEFINITION.md` | `PRD.md`, `tasks.yaml` (per-epic) |

The first three are portable expertise modules: they contain no host-specific file paths, no references to orchestrator control files, and no retry/loop semantics.

---

## Artifact Validation

`internal/state/validate.go` has a content validator for each managed artifact. `InferStage` calls validators during stage inference — a file that exists but contains only an unfilled shell is treated the same as a missing file. This prevents the pipeline from skipping a stage if the host pre-created empty shells before agent invocation.

---

## Key Design Decisions

**Artifact-derived position**: No separate state file. Pipeline position is always inferred from which artifacts are on disk and whether they are valid. This makes the state model simple and resilient — any time you re-run, the correct stage is inferred automatically.

**One step per `Run` call**: `orchestrator.Run` executes exactly one pipeline step and then returns. The user-facing loop is re-entry-based: each `doug-plan run` call resumes from the next inferred stage. This makes the orchestrator straightforward to test: give it a fake agent, call `Run`, inspect which artifact was written.

**Atomic file writes**: All file writes go to `<path>.tmp` first, then `os.Rename` to the final path. Prevents partial writes from corrupting artifacts if the process is killed mid-write.

**Dynamic port for the HTTP server**: `net.Listen("tcp", "127.0.0.1:0")` lets the OS pick a free port. No hardcoded port, no port-conflict failures. The URL is always printed to the terminal.

**No shell eval in `exec.Command`**: Agent commands and git operations always use an explicit args slice. No `sh -c`, no string concatenation. This is a hard rule that prevents command injection.

**Stdlib-only runtime**: Two external Go dependencies (`cobra`, `yaml.v3`). No logging library, no go-git, no alternative YAML parsers. Build-time Node.js for the UI bundle; the binary itself has no Node.js dependency at runtime.

---

## Adding a New Stage

1. Add a `Stage` constant to `internal/state/state.go` and update `StageFromString` and `ArtifactFile`.
2. Add a content validator in `internal/state/validate.go`.
3. Add an `ACTIVE_STEP.md` template in `internal/templates/steps/<Stage>.md` (optional; a generic template is used if absent).
4. Add an artifact shell template in `internal/templates/artifacts/<Artifact>` if the host should pre-create the file.
5. Wire the stage into `orchestrator.Run` if it requires special dispatch (e.g. the deterministic Handoff path).
6. Add a per-stage view to the React app (`ui/src/index.jsx`) and rebuild the bundle.

---

## Testing

```bash
make test                    # go test ./...
make test-integration        # go test -tags=integration ./...  (requires network/subprocess)
```

Unit tests are fast and do not open browsers or spawn real agent subprocesses — seams are injected via function parameters. Integration tests are tagged with `//go:build integration` and skipped by `go test -short`.

Table-driven tests are the preferred style throughout the codebase.
