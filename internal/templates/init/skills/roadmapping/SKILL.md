---
name: "roadmapping"
description: "Transform a VISION.md into a sequenced ROADMAP.md in the hybrid Markdown + YAML frontmatter format. Use when VISION.md exists and a roadmap of epics needs to be produced."
---

# Roadmapping Workflow

This skill reads `VISION.md` and produces a `ROADMAP.md` containing sequenced epics scoped at the "what are we building" level. It can be invoked standalone (`/roadmapping`) or as part of the `doug-plan` pipeline.

## Phase 1: Ingest Context

1. If `.doug/plan/ACTIVE_STEP.md` exists, read it for the planning brief.
2. Locate `VISION.md`:
   - Check `.doug/plan/VISION.md` first.
   - If not found, check the project root for `VISION.md`.
   - If neither exists, stop and tell the user that `VISION.md` is required before roadmapping can proceed.
3. Read `VISION.md` in full. Note the project name, goals, non-goals, constraints, and success criteria — these directly inform epic scope and sequencing.
4. If `.doug/plans/research/` exists, list its contents. Read every `.md` file found there. If the directory does not exist or is empty, proceed without it.

## Phase 2: Synthesize Epics

Using the vision and any research context, identify the minimal set of epics needed to achieve the stated goals. Apply these rules:

- **Scope each epic at the "what are we building" level** — not tasks, not user stories, not implementation details.
- **Sequence by dependency**: foundational infrastructure before features that depend on it; core user-facing value before enhancements.
- **Respect non-goals and constraints**: do not include epics for explicitly out-of-scope work.
- **Size for coherence**: each epic should deliver a meaningful, independently shippable increment. Avoid epics that are either trivially small or so large they cannot be reviewed as a unit.
- Aim for 3–8 epics. Fewer is better when scope allows.

For each epic, produce:
- A short, descriptive name (3–6 words)
- A one-paragraph description of what it builds and why it comes at this point in the sequence
- An EPIC-N identifier assigned in sequence order (EPIC-1, EPIC-2, …)

## Phase 3: Draft ROADMAP.md

Using the synthesized epics, draft a `ROADMAP.md` in the hybrid Markdown + YAML frontmatter format shown below.

**Top-level frontmatter** captures document metadata. **Each epic section** contains an embedded YAML block (parseable by the orchestrator) followed by a Markdown prose description.

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

One paragraph describing what this epic builds and why it comes first in the sequence.

## EPIC-2: Next Epic Title

---
id: EPIC-2
name: "Next Epic Title"
sequence: 2
status: planned
---

One paragraph describing what this epic builds and why it follows EPIC-1.
```

Rules for the output document:
- Every epic must have a complete YAML block with `id`, `name`, `sequence`, and `status` fields.
- `status` is always `planned` for newly produced roadmaps.
- The `sequence` values must be consecutive integers starting at 1.
- Prose descriptions must be concrete — no placeholders, no "TBD", no bracketed text.
- Do not include epics for work that is explicitly out of scope in `VISION.md`.

## Phase 4: Review and Confirm

1. Present the full draft to the user.
2. Ask: "Does this roadmap correctly sequence the work? Any epics to add, remove, reorder, or rename before I save it?"
3. Apply any requested changes.
4. Repeat until the user confirms the roadmap is complete and correctly sequenced.

## Phase 5: Write Output

1. Ensure the directory `.doug/plan/` exists; create it if needed.
2. Write the confirmed document to `.doug/plan/ROADMAP.md`.
3. If `.doug/plan/ACTIVE_STEP.md` exists (pipeline mode), write the outcome into its `## Agent Result` block:

```markdown
## Agent Result

---
outcome: "SUCCESS"
---
```

4. Confirm to the user that `.doug/plan/ROADMAP.md` has been written.
