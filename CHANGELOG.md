# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- feat: EPIC-5-004 — scoped epic definition view with guided per-task fields; PRD/tasks.yaml split layout view
- feat: EPIC-5-003 — VISION.md and ROADMAP.md form views verified complete
- feat: EPIC-5-002 — build and embed compiled React bundle as self-contained HTML
- feat: EPIC-5-001 — embedded HTTP server with dynamic port, embed.FS bundle, browser gate
- feat: EPIC-4-005 — add Handoff ACTIVE_STEP.md template and full-pipeline e2e test
- feat: add greenfield detection and manifest.yaml emission to Handoff skill
- feat: implement Handoff skill for converting scoped epics to PRD.md and tasks.yaml
- feat: EPIC-4-002 — implement re-entry at Scoping for subsequent epics
- feat: implement Scoping skill and add StageScoping to pipeline
- feat: EPIC-3-005 — end-to-end pipeline validation through Roadmapping output; add orchestrator e2e test with fakeagent helper, ROADMAP.md format validator, and regression test for inline section reference bug
- feat: EPIC-3-004 — add ACTIVE_STEP.md templates for Discovery and Roadmapping stages
- feat: add roadmapping skill template for VISION.md → ROADMAP.md workflow in hybrid Markdown + YAML frontmatter format
- feat: add optional research report ingestion to Discovery skill
- feat: add discovery skill template for guided interview → VISION.md workflow
- feat: add golangci-lint configuration and refactor output formatting

### Changed
- docs: update KB for EPIC-4 — add Scoping/PRD stages and scoping/handoff skills
- docs(kb): update KB articles for EPIC-3 — document discovery and roadmapping skills, stage-specific ACTIVE_STEP.md templates, and ParseResult anchor fix
- fix: remove obsolete skill documentation and update scaffold tests to use research skill

### Fixed
- fix: parseOutcome searched for "## Agent Result" as a substring, matching the inline reference in Briefing text before the actual section heading; changed to line-level search ("\n## Agent Result")

### Removed

## [0.1.0]

### Added
- feat: scaffold embedded init templates for AGENTS.md, CLAUDE.md, provider configs, and default provider skill files
- feat: implement re-entry logic (resume, re-run, start fresh)
- feat: implement terminal approval gate (auto/soft/hard modes)
- feat: agent invocation, result parsing, outcome dispatch
- feat: implement ACTIVE_STEP.md write/read/archive lifecycle
- feat: artifact-presence state detection
- feat: add GitHub Actions release workflow for cross-platform binaries (macOS arm64, macOS amd64, Linux amd64)
- feat: update `doug-plan run` stub message to be clear and actionable
- feat: scaffold AGENTS.md and per-agent skill directories for claude, codex, and gemini
- Scaffold `.doug/plans/` directory, `ACTIVE_STEP.md` stub, and `doug-plan.yaml` config file with approval_mode, agent, and skill_paths fields.
- feat: implement `doug-plan init --agents` command with project scaffolding

### Changed
- refactor: move doug-plan-owned config and runtime files under `.doug/plan/` and add shared path helpers
- docs(kb): document EPIC-2 orchestrator loop, update project structure, add architecture section
- docs(kb): create README.md index and update infrastructure/go.md project structure

### Fixed
- fix: generate `approval_mode: auto` in init config so fresh scaffolds run without manual config repair
