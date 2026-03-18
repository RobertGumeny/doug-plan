# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- feat: implement terminal approval gate (auto/soft/hard modes)
- feat: EPIC-2-003 — agent invocation, result parsing, outcome dispatch
- feat: EPIC-2-002 — implement ACTIVE_STEP.md write/read/archive lifecycle
- feat: EPIC-2-001 — artifact-presence state detection
- feat: add GitHub Actions release workflow for cross-platform binaries (macOS arm64, macOS amd64, Linux amd64)
- feat: EPIC-1-004 — update `doug-plan run` stub message to be clear and actionable
- feat: scaffold AGENTS.md and per-agent skill directories for claude, codex, and gemini
- Scaffold `.doug/plans/` directory, `ACTIVE_STEP.md` stub, and `doug-plan.yaml` config file with approval_mode, agent, and skill_paths fields.
- feat: implement `doug-plan init --agents` command with project scaffolding

### Changed
- docs(kb): create README.md index and update infrastructure/go.md project structure

### Fixed

### Removed
