# Active Step

**Stage**: Discovery
**Artifact**: `.doug/plan/VISION.md`

## Briefing

Invoke `/discovery` to conduct a guided interview and produce `.doug/plan/VISION.md`.

**Target artifact**: `.doug/plan/VISION.md` — a shell has been pre-created with the required headings. Fill in each section; do not add or remove top-level headings.

**Required headings** (in order):
- `## Project Name`
- `## Problem Statement`
- `## Target Users`
- `## Goals`
- `## Non-Goals`
- `## Constraints`
- `## Success Criteria`
- `## Failure Conditions`
- `## Background`

**Supporting files to read** (if present):
- `.doug/plans/research/` — any prior research reports (optional)

The skill will:

1. Ingest any research reports from `.doug/plans/research/` (if present).
2. Ask an explicit greenfield vs. existing question to determine `project_mode`.
3. Ask structured questions to capture the project name, problem statement, target users, goals, non-goals, constraints, success criteria, and background.
4. **If greenfield**: collect scaffold inputs (language, runtime, framework, package manager, build system, runtime dependencies, dev dependencies, bootstrap constraints) and populate the VISION.md frontmatter accordingly.
5. **If existing**: leave scaffold frontmatter fields empty — they are not required for existing projects.
6. Draft `VISION.md` with the appropriate frontmatter and body, confirm it with the user, and write it to `.doug/plan/VISION.md`.
7. Write the outcome into this file's `## Agent Result` block before exiting.

---

## Agent Result

---
outcome: ""
---

## Output
