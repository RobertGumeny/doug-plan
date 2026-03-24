# doug-plan

`doug-plan` is a CLI tool that orchestrates multi-agent planning workflows. It drives a four-stage pipeline — Discovery → Roadmapping → Definition → Handoff — where each stage invokes an AI agent to produce a structured artifact. Between stages, the tool optionally gates on human approval before advancing.

---

## Prerequisites

- **Go 1.26+** — [install Go](https://go.dev/dl/)
- **Node.js and npm** — required only to rebuild the browser UI bundle; the committed `bundle.html` means you can skip this for a normal build

Verify your Go version:

```bash
go version   # should output go1.26.x or higher
```

---

## Install

### Build from source

```bash
git clone https://github.com/robertgumeny/doug-plan.git
cd doug-plan
make build          # builds UI bundle then compiles the binary
```

The binary is written to `./doug-plan` in the repo root.

If you do not need to rebuild the UI bundle (it is already committed):

```bash
go build -o doug-plan .
```

### Download a release binary

Pre-built binaries for macOS (arm64, amd64) and Linux (amd64) are available on the [Releases](https://github.com/robertgumeny/doug-plan/releases) page.

---

## Quick Start

### 1. Initialize a project

Run `init` inside any project directory. Pass the agents you want to use with `--agents`:

```bash
doug-plan init --agents claude
```

Supported agents: `claude`, `codex`, `gemini`. Comma-separate multiple agents:

```bash
doug-plan init --agents claude,codex
```

`init` creates:

```
.doug/plan/
├── doug-plan.yaml      ← configuration file
AGENTS.md               ← agent instructions
CLAUDE.md               ← Claude-specific instructions (if claude selected)
.claude/skills/         ← skill files copied for Claude
```

### 2. Configure the agent

Open `.doug/plan/doug-plan.yaml` and confirm the agent and approval mode:

```yaml
agent: claude
approval_mode: auto
```

`approval_mode` controls what happens after each successful stage:

| Mode | Behavior |
| ---- | -------- |
| `auto` | Advance immediately with no prompt |
| `soft` | Print summary in terminal; press Enter to advance or type `skip` to stop |
| `hard` | Open a browser review UI; block until you click Approve |

### 3. Run the pipeline

```bash
doug-plan run
```

The pipeline runs one stage per call, advancing automatically until complete or until it reaches an approval gate. Call `run` repeatedly to drive the pipeline forward:

```bash
doug-plan run   # Discovery
doug-plan run   # Roadmapping
doug-plan run   # Definition
doug-plan run   # Handoff (deterministic, no agent invoked)
```

When all stages are done the pipeline reports completion.

---

## Full Pipeline

Each stage reads the previous stage's artifact and produces the next one. All artifacts live in `.doug/plan/`.

```
Discovery  →  VISION.md
    ↓
Roadmapping  →  ROADMAP.md
    ↓
Definition  →  DEFINITION.md
    ↓
Handoff  →  PRD.md + tasks.yaml   (deterministic, no agent)
```

### Stage details

**Discovery** — The agent conducts a structured interview and synthesizes answers into `VISION.md`. Sections: Project Name, Problem Statement, Target Users, Goals, Non-Goals, Constraints, Success Criteria, Failure Conditions, Background. After Discovery approval, if `VISION.md` declares `project_mode: greenfield` in its frontmatter, `manifest.yaml` is written to `.doug/plan/manifest.yaml` with the stack and dependency information. Non-greenfield projects skip this step.

**Roadmapping** — The agent reads `VISION.md` and produces `ROADMAP.md`, a sequence of 3–8 epics in hybrid Markdown + YAML frontmatter format.

**Definition** — The agent reads `VISION.md` and `ROADMAP.md`, identifies the next undefined epic, and produces `DEFINITION.md` with a task breakdown and acceptance criteria per task.

**Handoff** — A deterministic code renderer (no agent) reads `DEFINITION.md` and produces `PRD.md` and `tasks.yaml` ready for downstream execution. Per-epic files are written to `.doug/plan/epics/<EPIC-ID>/`.

### Pipeline resume

`doug-plan` always infers its position from artifacts on disk. If you stop mid-run, the next `run` call picks up where you left off.

To re-run from a specific stage (clearing it and all later artifacts):

```bash
doug-plan run --rerun Roadmapping
```

To start completely fresh:

```bash
doug-plan run --fresh
```

### Approval mode override

Override the config-file setting for a single run:

```bash
doug-plan run --approval hard
```

---

## Reference

```
doug-plan init [flags]
  --agents string   comma-separated list of agents (claude, codex, gemini)

doug-plan run [flags]
  --approval string   approval mode override: auto, soft, or hard
  --rerun   string    re-run from stage: Discovery, Roadmapping, Definition, PRD, or Tasks
  --fresh             clear all plan artifacts and begin at Discovery
```

---

## Development

```bash
make build-ui    # rebuild the React bundle (requires Node.js)
make build       # build-ui + go build
make test        # go test ./...
make lint        # gofmt + go vet + golangci-lint
make clean       # remove the compiled binary
```

See [ARCHITECTURE.md](ARCHITECTURE.md) for a code-level walkthrough of the system.
