# Active Step

**Stage**: Roadmapping
**Artifact**: `.doug/plan/ROADMAP.md`

## Briefing

Invoke `/roadmapping` to transform `.doug/plan/VISION.md` into a sequenced `ROADMAP.md` in hybrid Markdown + YAML frontmatter format.

**Prerequisite**: `.doug/plan/VISION.md` must exist. If it is missing, set `outcome` to `FAILURE` and stop.

The skill will:

1. Read `.doug/plan/VISION.md` and any research reports from `.doug/plans/research/` (if present).
2. Synthesize a minimal set of sequenced epics derived from the vision goals, non-goals, and constraints (aim for 3–8 epics).
3. Draft `ROADMAP.md` in hybrid Markdown + YAML frontmatter format, confirm it with the user, and write it to `.doug/plan/ROADMAP.md`.
4. Write the outcome into this file's `## Agent Result` block before exiting.

---

## Agent Result

---
outcome: ""
---

## Output
