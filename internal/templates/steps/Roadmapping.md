# Active Step

**Stage**: Roadmapping
**Artifact**: `.doug/plan/ROADMAP.md`

## Briefing

Invoke `/roadmapping` to transform `.doug/plan/VISION.md` into a sequenced `ROADMAP.md` in hybrid Markdown + YAML frontmatter format.

**Prerequisite**: `.doug/plan/VISION.md` must exist. If it is missing, set `outcome` to `FAILURE` and stop.

**Target artifact**: `.doug/plan/ROADMAP.md` — a shell has been pre-created with the required structure. Replace the shell content with the completed roadmap; the top-level YAML frontmatter and `# Roadmap` heading must be preserved.

**Required structure**:
- Top-level YAML frontmatter with `project`, `generated`, and `source: VISION.md` fields
- `# Roadmap` heading
- One `## EPIC-N: <Title>` section per epic, each with an embedded YAML block containing `id`, `name`, `sequence`, and `status: planned`

**Supporting files to read**:
- `.doug/plan/VISION.md` — required; provides goals, non-goals, and constraints
- `.doug/plans/research/` — any prior research reports (optional)

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
