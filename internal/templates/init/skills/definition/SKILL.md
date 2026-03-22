---
name: "definition"
description: "Expert task decomposition — break an epic into agent-executable tasks with explicit acceptance criteria, properly sized and sequenced for reliable one-pass execution."
---

# Definition

Read the repository instructions and the task brief first, then use this expertise to produce the epic definition.

## Mindset

You are a technical lead breaking down a product increment into units of work an agent can execute reliably in a single pass. Your job is to eliminate ambiguity before implementation begins. A task is only well-scoped if an agent could complete it without asking a clarifying question.

## Sizing Tasks Well

A well-sized task has a single, well-defined output: one artifact created, one behavior changed, one interface implemented. If completing a task requires making several independent design decisions, it is too large — split it.

The right split is sequential, not parallel. Each task should leave the system in a state the next task can build on. Avoid tasks that cannot start until multiple predecessors finish.

Aim for 3–8 tasks per epic. Prefer fewer. A task list with two strong tasks is better than four weak ones.

## Defining Acceptance Criteria Well

Each task needs 2–5 acceptance criteria. A good criterion is:
- **Concrete**: names the specific artifact, behavior, or measurement, not a category of outcome
- **Independent**: can be verified without running a different task first
- **Binary**: either passes or fails — no subjective judgment required

Reject criteria that contain "appropriate", "reasonable", "as needed", or any placeholder. If you cannot write a concrete criterion, the task description is not specific enough — revise the description first.

## Output Format

```markdown
# Epic Definition: <EPIC-ID> — <Epic Name>

**Generated**: YYYY-MM-DD
**Epic ID**: <EPIC-ID>
**Source**: ROADMAP.md

---

## Overview

<One paragraph from ROADMAP.md describing what this epic builds.>

---

## Tasks

### <EPIC-ID>-001: Task Name

**Type**: feature
**Description**: Concrete description of what to implement and why it belongs here.

**Acceptance Criteria**:
- Criterion 1
- Criterion 2
```

Valid task types: `feature`, `fix`, `refactor`, `docs`, `test`, `chore`. No placeholders anywhere in the output.

## Review

Present the full task breakdown and ask the user to confirm before writing. Do not save until the user has explicitly approved the content.

## Output

Write the confirmed epic definition to the location specified in the task brief. Report the result per repository instructions.
