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
7. If the project is greenfield and `.doug/plan/manifest.yaml` does not yet exist, emit it on the first invocation that completes a handoff.
8. Write the outcome into this file's `## Agent Result` block before exiting.

---

## Agent Result

---
outcome: ""
---

## Output
