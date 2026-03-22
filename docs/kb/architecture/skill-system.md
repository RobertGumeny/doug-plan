---
title: Skill System
updated: 2026-03-19
category: Architecture
tags: [skills, scaffold, agents, claude, codex, gemini, discovery, roadmapping, definition, handoff]
related_articles:
  - docs/kb/architecture/orchestrator.md
---

# Skill System

## Overview

Skills are Markdown files that give an AI agent a structured workflow for a specific task. During `doug init`, doug-plan copies a set of baseline skill files into each agent's skill directory so the agent can invoke them by name (e.g. `/research`).

---

## Skill File Format

Each skill lives at `<agent-dir>/skills/<skill-name>/SKILL.md` and starts with a YAML frontmatter block:

```markdown
---
name: "skill-name"
description: "One-line description the agent reads when deciding whether to invoke this skill."
---

# Skill Title

Phases, steps, and rules the agent follows when the skill is active.
```

The `name` field must match the directory name. The `description` is surfaced to the agent during skill selection.

---

## Scaffold Copy Behavior

Skill templates live in `internal/templates/init/skills/`. During `scaffold.Run`, every file under that directory is copied to each selected agent's skill directory.

**Source layout:**

```
internal/templates/init/skills/
├── discovery/
│   └── SKILL.md       ← template
├── handoff/
│   └── SKILL.md       ← template
├── research/
│   └── SKILL.md       ← template
├── roadmapping/
│   └── SKILL.md       ← template
└── definition/
    └── SKILL.md       ← template
```

**Destination layout after `doug init --agents claude,codex`:**

```
.claude/skills/discovery/SKILL.md
.claude/skills/handoff/SKILL.md
.claude/skills/research/SKILL.md
.claude/skills/roadmapping/SKILL.md
.claude/skills/definition/SKILL.md
.codex/skills/discovery/SKILL.md
.codex/skills/handoff/SKILL.md
.codex/skills/research/SKILL.md
.codex/skills/roadmapping/SKILL.md
.codex/skills/definition/SKILL.md
```

The mapping is handled by `selectedSkillDestinations` in `internal/scaffold/scaffold.go`. Agent directories:

| Agent  | Skill directory        |
| ------ | ---------------------- |
| claude | `.claude/skills/`      |
| codex  | `.codex/skills/`       |
| gemini | `.gemini/skills/`      |

Files are written atomically (write to `.tmp`, then rename). If a destination file already exists it is skipped and logged as "Skipped".

---

## Adding a New Skill

1. Create `internal/templates/init/skills/<skill-name>/SKILL.md` using the format above.
2. Use the `research` skill as the baseline — it demonstrates the standard phase structure (Clarify → Gather Context → Explore → Produce Output → Finalize).
3. Run `go test ./internal/scaffold/...` to confirm the scaffold tests pass.
4. The next `doug init` on a new project will automatically copy the skill to every selected agent directory.

For an existing initialized project, copy the file manually:

```
cp internal/templates/init/skills/<skill-name>/SKILL.md .claude/skills/<skill-name>/SKILL.md
```

---

## Baseline Skill: `research`

`research` is the reference implementation for how a skill should be structured. It covers:

- A read-only analysis workflow (safe to invoke without side effects)
- Clear phase boundaries (Clarify Scope → Gather Context → Explore → Archive → Write Report → Finalize)
- Concrete tool guidance (`Glob`, `Grep`, `Read`) within phases
- A prescribed output format with defined sections

When building a new skill, start by copying `research/SKILL.md` and adapting the phases and output format to the new workflow.

---

## Planning Skills: `discovery`, `roadmapping`, `definition`, and `handoff`

These four skills implement the first four stages of the `doug-plan` pipeline. All can be invoked standalone or as part of a pipeline run.

### `discovery`

Runs a structured interview with the user and synthesizes the answers into `VISION.md`.

**Phases**: Ingest Existing Context → Guided Interview → Draft VISION.md → Review and Confirm → Write Output

- Phase 1 reads `.doug/plan/ACTIVE_STEP.md` (if present) and any research reports from `.doug/plans/research/`.
- Phase 2 asks 10 structured questions covering project identity, users, scope, constraints, and success criteria. Follow-up questions are asked until all answers are concrete (no "TBD" or placeholders).
- Phase 5 writes `.doug/plan/VISION.md` and, if running in pipeline mode, sets `outcome: "SUCCESS"` in `ACTIVE_STEP.md`.

**Output**: `.doug/plan/VISION.md` with sections: Project Name, Problem Statement, Target Users, Goals, Non-Goals, Constraints, Success Criteria, Failure Conditions, Background.

### `roadmapping`

Reads `VISION.md` and produces a `ROADMAP.md` containing sequenced epics in hybrid Markdown + YAML frontmatter format.

**Phases**: Ingest Context → Synthesize Epics → Draft ROADMAP.md → Review and Confirm → Write Output

- Phase 1 reads `.doug/plan/ACTIVE_STEP.md`, locates `VISION.md` (checks `.doug/plan/VISION.md` then project root), and reads any research reports from `.doug/plans/research/`.
- Phase 2 derives a minimal set of 3–8 epics, scoped at the "what are we building" level, sequenced by dependency.
- Phase 5 writes `.doug/plan/ROADMAP.md` and, if running in pipeline mode, sets `outcome: "SUCCESS"` in `ACTIVE_STEP.md`.

**Output format** — hybrid Markdown + YAML frontmatter:

```markdown
---
project: "Project Name"
generated: "YYYY-MM-DD"
source: VISION.md
---

# Roadmap

## EPIC-1: Epic Title

---
id: EPIC-1
name: "Epic Title"
sequence: 1
status: planned
---

One paragraph description.
```

Each epic block must have `id`, `name`, `sequence`, and `status` fields. `status` is always `planned` for newly produced roadmaps.

### `definition`

Reads `VISION.md` and `ROADMAP.md`, identifies the next undefined epic, and produces a task breakdown with acceptance criteria.

**Phases**: Ingest Context → Identify Next Epic → Scope the Target Epic → Draft and Review → Write Output

- Phase 1 reads `.doug/plan/ACTIVE_STEP.md` (if present), locates `VISION.md` and `ROADMAP.md`.
- Phase 2 finds the first epic in ROADMAP.md that does not have `.doug/plan/epics/<EPIC-ID>/DEFINITION.md`.
- Phase 3 produces 3–8 tasks per epic, each with an ID (`<EPIC-ID>-NNN`), type, description, and 2–5 measurable acceptance criteria.
- Phase 5 writes `.doug/plan/epics/<EPIC-ID>/DEFINITION.md`. If all epics are now defined, it also writes `.doug/plan/DEFINITION.md` and sets `outcome` to `SUCCESS`. If more epics remain, it sets `outcome` to `RETRY` so the orchestrator re-invokes the skill for the next epic.

**Output**: Per-epic `.doug/plan/epics/<EPIC-ID>/DEFINITION.md` files, plus `.doug/plan/DEFINITION.md` when all epics are defined.

### `handoff`

Reads an epic definition and converts it into `PRD.md` and `tasks.yaml` conforming to the `doug` template format.

**Phases**: Ingest Context → Identify Next Epic to Hand Off → Convert to PRD.md and tasks.yaml → Draft and Review → Write Output

- Phase 1 reads `.doug/plan/ACTIVE_STEP.md` (if present), locates `VISION.md` and `ROADMAP.md`, and determines whether the project is **greenfield** (no references to existing codebase, legacy system, migration, or rewrite in `VISION.md`).
- Phase 2 finds the first epic that has `DEFINITION.md` but is missing either `PRD.md` or `tasks.yaml`.
- Phase 3 converts `DEFINITION.md` into a `PRD.md` (epic overview, goals, non-goals, background, success criteria, deliverables, acceptance criteria) and a `tasks.yaml` (task list with id, type, status, description, acceptance_criteria).
- Phase 5 writes both files to `.doug/plan/epics/<EPIC-ID>/`. When all defined epics are handed off, it writes `.doug/plan/PRD.md` and sets `outcome` to `SUCCESS`. If more remain, it sets `outcome` to `RETRY`.
- **Greenfield manifest**: On the first successful handoff for a greenfield project, emits `.doug/plan/manifest.yaml` with fields: `project`, `generated`, `greenfield`, `stack`, `build_system`, `dependencies`.

**Output**: Per-epic `.doug/plan/epics/<EPIC-ID>/PRD.md` and `tasks.yaml`, plus `.doug/plan/PRD.md` when all defined epics are handed off. Optionally `.doug/plan/manifest.yaml` for greenfield projects.

---

## Skill Invocation

Agents invoke skills via slash commands matching the skill `name` field (e.g. `/research`, `/discovery`, `/roadmapping`). The agent reads the `SKILL.md` content and follows its phases for the duration of that invocation.

Skills are agent-local: each agent directory gets its own copy. Changes to a skill template do not retroactively update existing initialized projects.
