# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- feat: EPIC-3-004 — add ACTIVE_STEP.md templates for Discovery and Roadmapping stages
- feat: add roadmapping skill template for VISION.md → ROADMAP.md workflow in hybrid Markdown + YAML frontmatter format
- feat: add optional research report ingestion to Discovery skill
- feat: add discovery skill template for guided interview → VISION.md workflow

### Changed

- fix: remove obsolete skill documentation and update scaffold tests to use research skill

### Fixed

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
