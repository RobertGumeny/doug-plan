---
name: "handoff"
description: "Convert a scoped epic definition into a doug-ready PRD.md and tasks.yaml. Use when SCOPED.md exists for an epic and PRD.md or tasks.yaml have not yet been written for it. Handles re-entry automatically — targets the next unhandled epic each time it is invoked."
---

# Handoff Workflow

This skill reads `.doug/plan/epics/<EPIC-ID>/SCOPED.md`, converts the scoped task definitions into `PRD.md` and `tasks.yaml` conforming to the `doug` template format, and writes both files to `.doug/plan/epics/<EPIC-ID>/`. It can be invoked standalone (`/handoff`) or as part of the `doug-plan` pipeline.

## Phase 1: Ingest Context

1. If `.doug/plan/ACTIVE_STEP.md` exists, read it for the planning brief.
2. Locate `VISION.md`:
   - Check `.doug/plan/VISION.md` first.
   - If not found, check the project root.
   - If neither exists, stop and tell the user that `VISION.md` is required before handoff can proceed.
3. Read `VISION.md` in full. Note the goals, non-goals, constraints, and success criteria — these inform the PRD background and scope.
4. Locate `ROADMAP.md`:
   - Check `.doug/plan/ROADMAP.md` first.
   - If not found, check the project root.
   - If neither exists, stop and tell the user that `ROADMAP.md` is required before handoff can proceed.
5. Read `ROADMAP.md` in full. Parse each epic block (id, name, sequence, description).
6. Determine whether the project is **greenfield**. A project is greenfield if ALL of the following hold:
   - `VISION.md` contains no references to an existing codebase, legacy system, migration, rewrite, or port.
   - The problem statement and background describe a new creation, not an extension or replacement of prior software.
   - No constraints in `VISION.md` mention integrating with or preserving a pre-existing implementation.
   Record this determination (greenfield: true/false) for use in Phase 5.

## Phase 2: Identify the Next Epic to Hand Off

1. For each epic in ROADMAP.md (in sequence order):
   - Check whether `.doug/plan/epics/<epic-id>/SCOPED.md` exists.
   - Check whether `.doug/plan/epics/<epic-id>/PRD.md` exists.
   - Check whether `.doug/plan/epics/<epic-id>/tasks.yaml` exists.
2. The **first epic that has `SCOPED.md` but is missing either `PRD.md` or `tasks.yaml`** is the target for this invocation.
3. If no such epic exists (either all scoped epics have both files, or no epics are scoped), stop and tell the user the current state.

## Phase 3: Convert the Scoped Epic to PRD.md and tasks.yaml

Read `.doug/plan/epics/<EPIC-ID>/SCOPED.md` in full. Use its task definitions, acceptance criteria, and overview — combined with context from `VISION.md` and `ROADMAP.md` — to produce both output files.

### PRD.md format

```markdown
# PRD — <EPIC-ID>: <Epic Name>

## Epic

- id: <EPIC-ID>
- name: <Epic Name>

## Overview

<One paragraph describing what this epic builds and why it matters, drawn from ROADMAP.md and SCOPED.md.>

## Scope

In-scope: <Comma-separated list of the concrete deliverables in this epic.>

Out-of-scope: <Explicit exclusions — items that are deferred, belong to adjacent epics, or are not required for this epic's success.>

## Goals

- Goal 1: <Measurable outcome this epic achieves.>
- Goal 2: <Additional measurable outcome, if applicable.>

## Non-Goals

- <Item explicitly excluded from this epic.>

## Background / Context

<1–2 sentences of background connecting this epic to the project vision and the epics that precede it.>

## Success Criteria

- Criterion A: <Concrete, independently verifiable success condition.>
- Criterion B: <Additional criterion, if applicable.>

## Deliverables

- <Deliverable 1 — maps to one or more tasks in SCOPED.md.>
- <Deliverable 2>

## Acceptance Criteria

- Acceptance 1: <Directly derived from the task acceptance criteria in SCOPED.md.>
- Acceptance 2: <Additional criterion.>

## Notes for Agents

Refer to AGENTS.md for further instructions and `docs/kb` for additional context around project structure and best practices
```

**Rules for PRD.md:**
- No placeholders, no "TBD", no bracketed text in the final output.
- Goals and Success Criteria must be measurable.
- Acceptance Criteria must be independently verifiable — copy and refine them directly from the task acceptance criteria in SCOPED.md.
- Out-of-scope must list at least one explicit exclusion.

### tasks.yaml format

```yaml
epic:
  id: "<EPIC-ID>"
  name: "<Epic Name>"
  tasks:
    - id: "<EPIC-ID>-001"
      type: "feature"
      status: "TODO"
      description: "<1–3 sentences describing what to implement. Must be concrete — no placeholders.>"
      acceptance_criteria:
        - "<Criterion 1>"
        - "<Criterion 2>"
    - id: "<EPIC-ID>-002"
      type: "feature"
      status: "TODO"
      description: "<Description.>"
      acceptance_criteria:
        - "<Criterion 1>"
        - "<Criterion 2>"
```

**Rules for tasks.yaml:**
- `id` values must match the task IDs in SCOPED.md exactly (e.g., `EPIC-1-001`).
- `type` must be one of: `feature`, `fix`, `refactor`, `docs`, `test`, `chore`.
- `status` is always `"TODO"` in the initial output.
- `description` must be a single string (no nested YAML structures).
- `acceptance_criteria` items must be directly derived from the acceptance criteria in SCOPED.md with no placeholders.
- All string values must be double-quoted.

## Phase 4: Draft and Review

Present both files to the user:

1. Show `PRD.md` first with the heading `## PRD.md`.
2. Show `tasks.yaml` below with the heading `## tasks.yaml`.

Ask the user: "Does this PRD and task list correctly represent <EPIC-ID>? Any changes before I write the files?"

Apply any requested changes and repeat until the user confirms the output is ready.

## Phase 5: Write Output

1. Ensure the directory `.doug/plan/epics/<EPIC-ID>/` exists; create it if needed.
2. Write the confirmed content to `.doug/plan/epics/<EPIC-ID>/PRD.md`.
3. Write the confirmed content to `.doug/plan/epics/<EPIC-ID>/tasks.yaml`.
4. Check whether all epics in ROADMAP.md that have a `SCOPED.md` now also have both `PRD.md` and `tasks.yaml`:
   - **All scoped epics handed off**: Write `.doug/plan/PRD.md` with the content below, then set `outcome` to `SUCCESS`.
   - **More scoped epics remain**: Do not write `.doug/plan/PRD.md`. Set `outcome` to `RETRY` so the orchestrator re-invokes this skill for the next epic.

**`.doug/plan/PRD.md` content (written only when all scoped epics are handed off):**

```markdown
# Handoff Complete

**Generated**: YYYY-MM-DD

All scoped epics have been handed off. Per-epic PRD.md and tasks.yaml files are in `.doug/plan/epics/`.
```

5. If `.doug/plan/ACTIVE_STEP.md` exists (pipeline mode), write the outcome into its `## Agent Result` block:

```markdown
## Agent Result

---
outcome: "SUCCESS"
---
```

Use `"RETRY"` instead of `"SUCCESS"` when more epics remain to be handed off.

6. **Greenfield manifest**: If the project was determined to be greenfield in Phase 1 AND `.doug/plan/manifest.yaml` does not already exist, emit `.doug/plan/manifest.yaml` using the format below. If `.doug/plan/manifest.yaml` already exists, skip this step.

**`.doug/plan/manifest.yaml` format:**

```yaml
project: "<Project Name from VISION.md>"
generated: "YYYY-MM-DD"
greenfield: true
stack:
  - "<primary language or runtime, e.g. Go>"
  - "<additional language, framework, or platform if applicable>"
build_system: "<primary build tool, e.g. make, gradle, npm, cargo, go>"
dependencies:
  - "<notable runtime dependency>"
  - "<additional dependency if applicable>"
```

**Rules for `manifest.yaml`:**
- `project` must match the project name in `VISION.md` exactly.
- `stack` must list every language, runtime, and major framework mentioned in `VISION.md` or `ROADMAP.md`. Each entry is a plain string. No empty entries.
- `build_system` must be a single string identifying the primary build tool. Derive it from technology choices in `VISION.md`; if not specified, use the conventional default for the stack (e.g. `go` for Go-only projects, `npm` for Node.js).
- `dependencies` lists notable runtime dependencies derived from `VISION.md` or `ROADMAP.md`. If none are specified, use an empty list (`[]`).
- No placeholders, no "TBD", no bracketed text. Every field must be unambiguous.
- Do **not** emit `manifest.yaml` if the project is not greenfield.

7. Confirm to the user which epic was handed off, whether more remain, and (if applicable) whether `manifest.yaml` was written.
