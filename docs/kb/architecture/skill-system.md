---
title: Skill System
updated: 2026-03-18
category: Architecture
tags: [skills, scaffold, agents, claude, codex, gemini]
related_articles:
  - docs/kb/architecture/orchestrator.md
---

# Skill System

## Overview

Skills are Markdown files that give an AI agent a structured workflow for a specific task. During `doug init`, doug-plan copies a set of baseline skill files into each agent's skill directory so the agent can invoke them by name (e.g. `/research`).

---

## Skill File Format

Each skill lives at `<agent-dir>/skills/<skill-name>/SKILL.md` and starts with a YAML frontmatter block:

```markdown
---
name: "skill-name"
description: "One-line description the agent reads when deciding whether to invoke this skill."
---

# Skill Title

Phases, steps, and rules the agent follows when the skill is active.
```

The `name` field must match the directory name. The `description` is surfaced to the agent during skill selection.

---

## Scaffold Copy Behavior

Skill templates live in `internal/templates/init/skills/`. During `scaffold.Run`, every file under that directory is copied to each selected agent's skill directory.

**Source layout:**

```
internal/templates/init/skills/
└── research/
    └── SKILL.md       ← template
```

**Destination layout after `doug init --agents claude,codex`:**

```
.claude/skills/research/SKILL.md
.codex/skills/research/SKILL.md
```

The mapping is handled by `selectedSkillDestinations` in `internal/scaffold/scaffold.go`. Agent directories:

| Agent  | Skill directory        |
| ------ | ---------------------- |
| claude | `.claude/skills/`      |
| codex  | `.codex/skills/`       |
| gemini | `.gemini/skills/`      |

Files are written atomically (write to `.tmp`, then rename). If a destination file already exists it is skipped and logged as "Skipped".

---

## Adding a New Skill

1. Create `internal/templates/init/skills/<skill-name>/SKILL.md` using the format above.
2. Use the `research` skill as the baseline — it demonstrates the standard phase structure (Clarify → Gather Context → Explore → Produce Output → Finalize).
3. Run `go test ./internal/scaffold/...` to confirm the scaffold tests pass.
4. The next `doug init` on a new project will automatically copy the skill to every selected agent directory.

For an existing initialized project, copy the file manually:

```
cp internal/templates/init/skills/<skill-name>/SKILL.md .claude/skills/<skill-name>/SKILL.md
```

---

## Baseline Skill: `research`

`research` is the reference implementation for how a skill should be structured. It covers:

- A read-only analysis workflow (safe to invoke without side effects)
- Clear phase boundaries (Clarify Scope → Gather Context → Explore → Archive → Write Report → Finalize)
- Concrete tool guidance (`Glob`, `Grep`, `Read`) within phases
- A prescribed output format with defined sections

When building a new skill, start by copying `research/SKILL.md` and adapting the phases and output format to the new workflow.

---

## Skill Invocation

Agents invoke skills via slash commands matching the skill `name` field (e.g. `/research`). The agent reads the `SKILL.md` content and follows its phases for the duration of that invocation.

Skills are agent-local: each agent directory gets its own copy. Changes to a skill template do not retroactively update existing initialized projects.
