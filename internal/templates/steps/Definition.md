# Active Step

**Stage**: Definition
**Artifact**: `.doug/plan/DEFINITION.md`

## Briefing

Invoke `/definition` to define the next unmaterialized epic from `.doug/plan/ROADMAP.md`.

**Prerequisites**: `.doug/plan/VISION.md` and `.doug/plan/ROADMAP.md` must exist. If either is missing, set `outcome` to `FAILURE` and stop.

**Target artifact**: `.doug/plan/DEFINITION.md` — a shell has been pre-created with the required headings. Write the completed definition index here only when all epics are defined; on RETRY leave this file untouched (the host will remove it and recreate the shell on the next invocation).

**Required headings for** `.doug/plan/DEFINITION.md`:
- `# Definition`
- `## Defined Epics`

The per-epic DEFINITION.md schema is the authoritative source for downstream rendering. Each section maps directly to the output PRD.md (Overview → Overview, Scope → Scope, Goals → Goals, Non-Goals → Non-Goals, Background → Background, Success Criteria → Success Criteria, Deliverables → Deliverables) and the Tasks section maps to tasks.yaml.

**Per-epic artifact**: `.doug/plan/epics/<EPIC-ID>/DEFINITION.md` — write one file per epic as it is defined. Required headings:
- YAML frontmatter block: `id: "<EPIC-ID>"` and `name: "<Epic Name>"`
- `# Definition`
- `## Overview`
- `## Scope` (with `### In-Scope` and `### Out-of-Scope` subsections)
- `## Goals`
- `## Non-Goals`
- `## Background`
- `## Success Criteria`
- `## Deliverables`
- `## Tasks` (with `### <EPIC-ID>-NNN: <Task Name>` subsections, each including `**Type**`, `**Description**`, and `**Acceptance Criteria**` bullet list)

**Supporting files to read**:
- `.doug/plan/VISION.md` — required; provides project context
- `.doug/plan/ROADMAP.md` — required; provides the epic list and sequencing
- `.doug/plan/epics/*/DEFINITION.md` — read existing epic definitions to determine which epic is next

The skill will:

1. Read `.doug/plan/VISION.md` and `.doug/plan/ROADMAP.md`.
2. Identify the first epic in ROADMAP.md that does not yet have `.doug/plan/epics/<EPIC-ID>/DEFINITION.md`.
3. Break the epic into tasks sized for reliable agent execution, each with explicit acceptance criteria.
4. Present the epic definition to the user, apply any corrections, and write it to `.doug/plan/epics/<EPIC-ID>/DEFINITION.md`.
5. If all epics are now defined, write `.doug/plan/DEFINITION.md` and set `outcome` to `SUCCESS`.
6. If more epics remain, set `outcome` to `RETRY` — the orchestrator will re-invoke this skill for the next epic.
7. Write the outcome into this file's `## Agent Result` block before exiting.

---

## Agent Result

---
outcome: ""
---

## Output
