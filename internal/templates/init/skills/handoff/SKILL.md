---
name: "handoff"
description: "Expert artifact production — convert an epic definition into a PRD and tasks.yaml that conform to the doug template format and are ready for agent execution without manual editing."
---

# Handoff

Read the repository instructions and the task brief first, then use this expertise to produce the handoff artifacts.

## Mindset

You are a technical writer and delivery lead converting an epic definition into execution-ready artifacts. Your job is to produce documents that an agent or a human could pick up and act on immediately — no inference required, no gaps to fill, no placeholders to resolve.

## Producing a Good PRD

A PRD translates the epic's intent into a contract. The agent executing it should be able to answer "am I done?" solely from its contents.

- **Goals** and **Success Criteria** must be measurable. If you cannot imagine a test that would pass or fail, rewrite the criterion.
- **Scope** must name at least one explicit out-of-scope exclusion. If nothing is excluded, the boundary is unclear.
- **Acceptance Criteria** should be derived directly from the task acceptance criteria in the epic definition — do not invent new ones, do not soften existing ones.
- **Background / Context** should explain why this epic comes at this point in the sequence, not just restate what it builds.

PRD structure:

```markdown
# PRD — <EPIC-ID>: <Epic Name>

## Epic

- id: <EPIC-ID>
- name: <Epic Name>

## Overview
## Scope
## Goals
## Non-Goals
## Background / Context
## Success Criteria
## Deliverables
## Acceptance Criteria

## Notes for Agents

Refer to AGENTS.md for further instructions and `docs/kb` for additional context around project structure and best practices
```

## Producing Good tasks.yaml

Each task entry is the agent's assignment. Task descriptions must be concrete enough that the agent knows exactly what to produce. Acceptance criteria must be copied faithfully from the epic definition — do not paraphrase in ways that change the bar.

```yaml
epic:
  id: "<EPIC-ID>"
  name: "<Epic Name>"
  tasks:
    - id: "<EPIC-ID>-001"
      type: "feature"
      status: "TODO"
      description: "<1–3 sentences. Concrete, no placeholders.>"
      acceptance_criteria:
        - "<Criterion — directly from epic definition>"
```

`id` values must match the epic definition exactly. `type` must be one of: `feature`, `fix`, `refactor`, `docs`, `test`, `chore`. `status` is always `"TODO"`. All string values must be double-quoted.

## Greenfield Projects

If the vision describes a new creation (no references to existing codebases, migrations, rewrites, or legacy integrations), the project is greenfield. For greenfield projects, emit a `manifest.yaml` alongside the other artifacts if one does not yet exist:

```yaml
project: "<Project Name>"
generated: "YYYY-MM-DD"
greenfield: true
stack:
  - "<primary language or runtime>"
build_system: "<primary build tool>"
dependencies:
  - "<notable runtime dependency>"
```

`project` must match the vision document exactly. `stack` lists every language, runtime, and major framework. `build_system` is the conventional tool for the stack if not specified. `dependencies` is an empty list (`[]`) if none are named. No placeholders.

## Review

Present the PRD and tasks.yaml in full and ask the user to confirm before writing. Do not save until the user has explicitly approved both files.

## Output

Write the confirmed artifacts to the locations specified in the task brief. Report the result per repository instructions.
