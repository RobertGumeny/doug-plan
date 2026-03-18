---
name: "discovery"
description: "Conduct a guided interview to capture project vision, goals, constraints, and success criteria, then synthesize the result into VISION.md. Use when starting a new project or when a vision document does not yet exist."
---

# Discovery Workflow

This skill runs a structured interview with the user and synthesizes the answers into a complete `VISION.md`. It can be invoked standalone (`/discovery`) or as part of the `doug-plan` pipeline.

## Phase 1: Ingest Existing Context

Before asking questions, gather any context already available:

1. If `.doug/plan/ACTIVE_STEP.md` exists, read it for the planning brief.
2. If `.doug/plans/research/` exists, list any files inside and read the ones most relevant to the project.
3. Note every piece of context that can pre-fill or inform interview answers.

## Phase 2: Guided Interview

Ask the following questions in order. If the user's initial message already answers a question, acknowledge the answer and skip or confirm it rather than re-asking. Do not proceed to Phase 3 until every question has a concrete, non-placeholder answer.

**Project identity**
1. What is the name of this project or product?
2. In one sentence, what problem does it solve and why does that problem matter?

**Users and goals**
3. Who are the primary users or customers?
4. What is the single most important outcome for those users?

**Scope**
5. What is explicitly in scope for the first version?
6. What is explicitly out of scope?

**Constraints**
7. Are there hard technical, legal, budget, or timeline constraints?

**Success and failure**
8. How will you know the project has succeeded? What measurable outcome signals "done"?
9. What outcomes must be avoided — what does failure look like?

**Background**
10. Is there prior work, existing systems, or related research to be aware of?

**Follow-up rule**: If any answer is vague, circular, or contains a placeholder (e.g., "TBD", "not sure", "to be decided"), ask a targeted follow-up question until the answer is concrete. Do not synthesize `VISION.md` while any answer remains unresolved.

## Phase 3: Draft VISION.md

Using the interview answers, draft a `VISION.md` with the structure below. Every section must contain concrete content. Do not leave any field blank, use "TBD", or include bracketed placeholders in the final document.

```markdown
# Vision

## Project Name

[Name of the project or product]

## Problem Statement

[One paragraph describing the problem, why it matters, and who is affected]

## Target Users

[Description of primary users or customers and what they need]

## Goals

- [Concrete goal 1]
- [Concrete goal 2]
- [Concrete goal 3]

## Non-Goals

- [What is explicitly out of scope for the first version]

## Constraints

- [Technical, legal, budget, or timeline constraints]

## Success Criteria

- [Measurable outcome that signals success]

## Failure Conditions

- [Outcomes that would constitute failure]

## Background

[Prior work, existing systems, research reports, or relevant context]
```

## Phase 4: Review and Confirm

1. Present the full draft to the user.
2. Ask: "Does this accurately capture your vision? Any corrections or additions before I save it?"
3. Apply any corrections.
4. Repeat until the user confirms the document is complete.

## Phase 5: Write Output

1. Ensure the directory `.doug/plan/` exists; create it if needed.
2. Write the confirmed document to `.doug/plan/VISION.md`.
3. If `.doug/plan/ACTIVE_STEP.md` exists (pipeline mode), write the outcome into its `## Agent Result` block:

```markdown
## Agent Result

---
outcome: "SUCCESS"
---
```

4. Confirm to the user that `.doug/plan/VISION.md` has been written.
