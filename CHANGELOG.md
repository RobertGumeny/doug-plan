# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Add session-completion instruction to all step templates and the generic fallback, directing agents to inform the user the step is complete and to exit and run `doug-plan run`.
- Harden outcome parser to normalize to uppercase before validation; add valid-values comment to all step templates and generic fallback.
- feat: hard-rename approval modes soft→cli and hard→browser across all code, config, CLI help, tests, and documentation
- Remove --print flag from default Claude agent command so doug-plan run opens an interactive shell session instead of a one-shot output.
- Add guided provider selection prompt to `doug-plan init` using `prompt.SelectOne`; TTY presents numbered menu, non-TTY silently defaults to claude, `--agents` flag bypasses prompt.
- feat: add internal/prompt package with SelectOne and Text helpers; TTY/non-TTY paths tested
- docs: remove stale manifest references from handoff skill and PRD template; document post-Discovery manifest sync in orchestrator.md and README
- Add tests for manifest generation and approval paths (AC: greenfield emit, existing-project no-manifest, rerun cleanup+regeneration, soft/auto malformed-frontmatter error, hard-mode malformed-frontmatter no-crash)
- feat: Discovery hard-approval UI updated to split view for greenfield projects (VISION.md + manifest.yaml); non-greenfield projects retain single-pane behaviour
- Integrate manifest.Sync into orchestrator post-Discovery flow: call Sync after Discovery approval in all modes, remove manifest on --fresh and Discovery rerun.
- feat(manifest): add Sync entry point — writes or removes .doug/plan/manifest.yaml based on VISION.md frontmatter; all-or-nothing atomic write; human-readable validation errors
- Update Discovery step brief and skill contract to capture machine-readable scaffold inputs for greenfield projects
- feat: define manifest schema v1 and extend VISION.md template with greenfield frontmatter
- docs: add README.md and ARCHITECTURE.md for public audience
- Add artifact validation helpers for all five managed artifacts (VISION.md, ROADMAP.md, DEFINITION.md, PRD.md, tasks.yaml) and update InferStage to treat invalid or incomplete artifacts the same as missing ones; change StageTasks artifact from TASKS.md to tasks.yaml to align with the deterministic handoff output.
- Replace agent-driven handoff stage with deterministic renderer that reads per-epic DEFINITION.md and produces PRD.md and tasks.yaml; browser approval step preserved.
- feat(EPIC-6-003): expand DEFINITION.md schema and update DefinitionView for richer structured editing
- feat: EPIC-6-002 — host-owned artifact shells and enriched ACTIVE_STEP.md briefs
- feat: rewrite discovery, roadmapping, and definition skills to be host-agnostic

### Changed
- docs: update ARCHITECTURE.md, KB orchestrator article, and README for EPIC-8 changes (remove --print from Claude default command, add internal/prompt to package map, document interactive init provider selection)
- docs: update KB articles (browser-ui, orchestrator) to reflect EPIC-7 Discovery split-view and manifest package
- docs: update KB articles (orchestrator, skill-system) to reflect EPIC-6 changes

### Fixed

### Removed

## [0.1.1]

### Added
- feat: POST-back-to-disk approval flow and hard mode browser gate verified; pipeline blocks until browser Approve; content written atomically to disk before pipeline advances; browser auto-opens with terminal URL fallback
- feat: epic definition view with guided per-task fields; PRD/tasks.yaml split layout view
- feat: VISION.md and ROADMAP.md form views verified complete
- feat: build and embed compiled React bundle as self-contained HTML
- feat: embedded HTTP server with dynamic port, embed.FS bundle, browser gate
- feat: add Handoff ACTIVE_STEP.md template and full-pipeline e2e test
- feat: add greenfield detection and manifest.yaml emission to Handoff skill
- feat: implement Handoff skill for converting defined epics to PRD.md and tasks.yaml
- feat: implement re-entry at Definition for subsequent epics
- feat: implement Definition skill and add StageDefinition to pipeline
- feat: end-to-end pipeline validation through Roadmapping output; add orchestrator e2e test with fakeagent helper, ROADMAP.md format validator, and regression test for inline section reference bug
- feat: add ACTIVE_STEP.md templates for Discovery and Roadmapping stages
- feat: add roadmapping skill template for VISION.md → ROADMAP.md workflow in hybrid Markdown + YAML frontmatter format
- feat: add optional research report ingestion to Discovery skill
- feat: add discovery skill template for guided interview → VISION.md workflow
- feat: add golangci-lint configuration and refactor output formatting
- feat: add `make test-integration` and move server, agent invoke, and orchestrator e2e coverage behind the tagged suite

### Changed
- docs: add Browser UI KB article and update orchestrator/go KB articles for EPIC-5
- docs: update KB for EPIC-4 — add Definition/PRD stages and definition/handoff skills
- docs(kb): update KB articles for EPIC-3 — document discovery and roadmapping skills, stage-specific ACTIVE_STEP.md templates, and ParseResult anchor fix

### Fixed
- fix: remove obsolete skill documentation and update scaffold tests to use research skill
- fix: split fast unit tests from integration coverage with `integration` build tags  
- fix: clear errcheck failures in server tests by checking response body close errors explicitly
- fix: parseOutcome searched for "## Agent Result" as a substring, matching the inline reference in Briefing text before the actual section heading; changed to line-level search ("\n## Agent Result")
- fix: add seams so unit tests no longer open browsers or spawn real subprocesses

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
