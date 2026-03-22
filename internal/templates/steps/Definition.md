# Active Step

**Stage**: Definition
**Artifact**: `.doug/plan/DEFINITION.md`

## Briefing

Invoke `/definition` to define the next unmaterialized epic from `.doug/plan/ROADMAP.md`.

**Prerequisites**: `.doug/plan/VISION.md` and `.doug/plan/ROADMAP.md` must exist. If either is missing, set `outcome` to `FAILURE` and stop.

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
