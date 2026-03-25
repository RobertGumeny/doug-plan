# Active Step

**Stage**: PRD
**Artifact**: `.doug/plan/PRD.md`

## Briefing

Invoke `/handoff` to convert all epic definitions into `PRD.md` and `tasks.yaml` files ready for `doug`.

**Prerequisites**: `.doug/plan/VISION.md`, `.doug/plan/ROADMAP.md`, and `.doug/plan/DEFINITION.md` must exist. If any are missing, set `outcome` to `FAILURE` and stop.

The skill will:

1. Read `.doug/plan/VISION.md`, `.doug/plan/ROADMAP.md`, and `.doug/plan/DEFINITION.md`.
2. Identify the first epic that has `.doug/plan/epics/<EPIC-ID>/DEFINITION.md` but is missing either `PRD.md` or `tasks.yaml`.
3. Convert the epic definition into `PRD.md` and `tasks.yaml` and write both to `.doug/plan/epics/<EPIC-ID>/`.
4. Present the draft to the user, apply any corrections, and confirm before writing.
5. If all defined epics are now handed off, write `.doug/plan/PRD.md` and set `outcome` to `SUCCESS`.
6. If more defined epics remain, set `outcome` to `RETRY` — the orchestrator will re-invoke this skill for the next epic.
7. Write the outcome into this file's `## Agent Result` block before exiting.
8. Tell the user the step is complete and that they should exit this session and run `doug-plan run` to continue.

---

## Agent Result

---
outcome: "" # Must be one of: SUCCESS | FAILURE | RETRY
---

## Output

## Session Completion

After writing the outcome into the `## Agent Result` block, send the user this message:

> This step is complete. Please exit this session and run `doug-plan run` to continue.
