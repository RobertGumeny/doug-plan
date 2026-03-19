---
name: "scoping"
description: "Transform a ROADMAP.md epic into a scoped definition with tasks sized for reliable agent execution and explicit acceptance criteria. Use when VISION.md and ROADMAP.md exist and the next epic needs to be scoped. Handles re-entry automatically — targets the next unscoped epic each time it is invoked."
---

# Scoping Workflow

This skill reads `VISION.md` and `ROADMAP.md`, identifies the next epic that has not yet been scoped, and produces a complete task breakdown with acceptance criteria in `.doug/plan/epics/<EPIC-N>/SCOPED.md`. It can be invoked standalone (`/scoping`) or as part of the `doug-plan` pipeline.

## Phase 1: Ingest Context

1. If `.doug/plan/ACTIVE_STEP.md` exists, read it for the planning brief.
2. Locate `VISION.md`:
   - Check `.doug/plan/VISION.md` first.
   - If not found, check the project root.
   - If neither exists, stop and tell the user that `VISION.md` is required before scoping can proceed.
3. Read `VISION.md` in full. Note the goals, non-goals, constraints, and success criteria — these inform task scope and acceptance criteria.
4. Locate `ROADMAP.md`:
   - Check `.doug/plan/ROADMAP.md` first.
   - If not found, check the project root.
   - If neither exists, stop and tell the user that `ROADMAP.md` is required before scoping can proceed.
5. Read `ROADMAP.md` in full. Parse each epic block (id, name, sequence, description).

## Phase 2: Identify the Next Epic to Scope

1. For each epic in ROADMAP.md (in sequence order), check whether `.doug/plan/epics/<epic-id>/SCOPED.md` exists.
2. The **first epic whose `SCOPED.md` does not exist** is the target for this invocation.
3. If every epic in ROADMAP.md already has a `SCOPED.md`, all epics are scoped — proceed directly to Phase 5 (Write Output) to emit the completion marker and set outcome to SUCCESS.

## Phase 3: Scope the Target Epic

Using the epic description from ROADMAP.md and the constraints from VISION.md, produce a task breakdown for the target epic.

**Task sizing rules:**
- Each task must be completable by a single agent invocation without human intervention beyond approval gates.
- A task should have a single, well-defined output artifact or behavior change.
- If a unit of work is too large for one agent pass, split it into sequential tasks.
- Aim for 3–8 tasks per epic. Fewer is better when scope allows.

**For each task, define:**
- **ID**: `<EPIC-ID>-NNN` (e.g., `EPIC-1-001`, `EPIC-1-002`, …)
- **Name**: Short imperative phrase (3–7 words)
- **Type**: One of `feature`, `fix`, `refactor`, `docs`, `test`, `chore`
- **Description**: 1–3 sentences describing what to implement and why it belongs at this point in the sequence
- **Acceptance Criteria**: 2–5 concrete, measurable criteria — no placeholders, no "TBD", no bracketed text. Each criterion must be independently verifiable.

**Sequencing rules:**
- Order tasks so each one can build on the previous without circular dependencies.
- Infrastructure and foundational tasks come before tasks that depend on them.

## Phase 4: Draft and Review

Draft the scoped epic in the format below, then present it to the user.

```markdown
# Scoped Epic: <EPIC-ID> — <Epic Name>

**Generated**: YYYY-MM-DD
**Epic ID**: <EPIC-ID>
**Source**: ROADMAP.md

---

## Overview

<One paragraph from ROADMAP.md describing what this epic builds and why it comes at this point in the sequence.>

---

## Tasks

### <EPIC-ID>-001: Task Name

**Type**: feature
**Description**: Concrete description of what to implement.

**Acceptance Criteria**:
- Criterion 1
- Criterion 2

---

### <EPIC-ID>-002: Task Name

**Type**: feature
**Description**: Concrete description of what to implement.

**Acceptance Criteria**:
- Criterion 1
- Criterion 2

---
```

Ask the user: "Does this task breakdown correctly scope EPIC-N? Any tasks to add, remove, reorder, or rename before I save it?"

Apply any requested changes and repeat until the user confirms the scoped epic is complete.

## Phase 5: Write Output

1. Ensure the directory `.doug/plan/epics/<EPIC-ID>/` exists; create it if needed.
2. Write the confirmed document to `.doug/plan/epics/<EPIC-ID>/SCOPED.md`.
3. Check whether all epics in ROADMAP.md now have a `SCOPED.md` in their respective directories:
   - **All epics scoped**: Write `.doug/plan/SCOPED.md` with the content below, then set `outcome` to `SUCCESS`.
   - **More epics remain**: Do not write `.doug/plan/SCOPED.md`. Set `outcome` to `RETRY` so the orchestrator re-invokes this skill for the next epic.

**`.doug/plan/SCOPED.md` content (written only when all epics are scoped):**

```markdown
# Scoping Complete

**Generated**: YYYY-MM-DD

All epics from ROADMAP.md have been scoped. Scoped definitions are in `.doug/plan/epics/`.
```

4. If `.doug/plan/ACTIVE_STEP.md` exists (pipeline mode), write the outcome into its `## Agent Result` block:

```markdown
## Agent Result

---
outcome: "SUCCESS"
---
```

Use `"RETRY"` instead of `"SUCCESS"` when more epics remain to be scoped.

5. Confirm to the user which epic was scoped and whether more remain.
