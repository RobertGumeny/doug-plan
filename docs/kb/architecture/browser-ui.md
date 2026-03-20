---
title: Browser UI
updated: 2026-03-20
category: Architecture
tags: [browser, ui, server, approval, react, embed]
related_articles:
  - docs/kb/architecture/orchestrator.md
  - docs/kb/infrastructure/go.md
---

# Browser UI

## Overview

EPIC-5 replaces the terminal hard-approval gate with a purpose-built browser review experience. When approval mode is `hard`, the orchestrator starts an embedded HTTP server, opens the browser, serves a compiled React bundle, and blocks until the user approves. The approved (possibly edited) content is written back to disk; then the server shuts down and the pipeline advances.

The server spins up **per review step only** — it is not a persistent background process.

## Approval Flow

```
agent writes artifact → orchestrator calls BrowserGate
       │
       ▼
server.Serve: listen on 127.0.0.1:<random-port>
       │
       ├─ print "Review URL: http://127.0.0.1:<port>" to terminal
       ├─ openBrowser(url)  ← best-effort; failures are silent
       │
       ▼
browser loads GET /  → serves bundle.html (embedded React app)
browser loads GET /artifact → JSON: {stage, content, secondaryContent}
       │
       ▼
user reviews / edits in browser → clicks Approve
       │
       ▼
POST /approve {content, secondaryContent}
       │
       ├─ write content → artifactPath (atomic tmp→rename)
       ├─ write secondaryContent → secondaryPath if non-empty
       │
       ▼
server.Shutdown → BrowserGate returns nil → pipeline advances
```

## Hard Mode Dispatch (Orchestrator)

`runApprovalGate` in `internal/orchestrator` routes to `BrowserGate` only when:
- mode is `hard`, AND
- `state.ArtifactFile(stage)` returns a non-empty filename.

For the PRD stage, it also checks whether `tasks.yaml` exists alongside `PRD.md` and passes it as `secondaryPath` if present.

```go
if mode == approval.ModeHard {
    artifactFile := state.ArtifactFile(stage)
    if artifactFile != "" {
        primaryPath := filepath.Join(plansDir, artifactFile)
        secondaryPath := ""
        if stage == state.StagePRD {
            // tasks.yaml is served alongside PRD.md when present
            candidate := filepath.Join(plansDir, "tasks.yaml")
            if _, statErr := os.Stat(candidate); statErr == nil {
                secondaryPath = candidate
            }
        }
        return approval.BrowserGate(primaryPath, secondaryPath, stage.String(), opts.Out)
    }
}
```

`auto` and `soft` modes still use the terminal gate unchanged.

## HTTP API

| Method | Path       | Description |
| ------ | ---------- | ----------- |
| `GET`  | `/`        | Serves the self-contained `bundle.html` |
| `GET`  | `/artifact`| Returns `{stage, content, secondaryContent}` as JSON |
| `POST` | `/approve` | Accepts `{content, secondaryContent}`; writes to disk; shuts down server |

The server listens on `127.0.0.1:0` (OS-assigned dynamic port). The full URL is always printed to the terminal as a fallback.

## React Bundle & Views

The React app is a single self-contained HTML file (`internal/ui/bundle.html`) with no runtime external dependencies. It is embedded via Go's `embed.FS` in `internal/ui/ui.go`.

Views are rendered based on the `stage` field returned by `GET /artifact`:

| Stage | View | Description |
| ----- | ---- | ----------- |
| `Discovery` | `VisionView` | Single large textarea for `VISION.md` |
| `Roadmapping` | `RoadmapView` | Epic card list with editable title, description, reordering |
| `Scoping` | `ScopingView` | Task list with guided fields per task |
| `PRD` | `PRDView` | Split prose textarea and structured task list |
| `Tasks` | `TasksView` | Structured task review |

## Build Pipeline

The bundle must be compiled before the Go binary. `make build` handles both steps:

```
make build-ui   →  npm install --prefix ui
                   node ui/build.js
                   → writes internal/ui/bundle.html

make build      →  (runs build-ui first)
                   go build -o doug-plan .
```

`ui/build.js` uses `esbuild` to bundle `ui/src/index.jsx` as an IIFE, minifies it, and wraps it in a minimal HTML shell. The output is written directly to `internal/ui/bundle.html` and committed. Node.js and npm are **build-time** dependencies only — the runtime binary has no Node.js dependency.

## Package Map

| Package | Responsibility |
| ------- | -------------- |
| `internal/server` | `Serve` — HTTP server, endpoint handlers, browser open, atomic file writes |
| `internal/ui` | `Bundle embed.FS` — holds the compiled `bundle.html` |
| `internal/approval` | `BrowserGate` — thin wrapper that delegates to `server.Serve` |

## Key Decisions

**Dynamic port**: `net.Listen("tcp", "127.0.0.1:0")` lets the OS pick a free port. No hardcoded port, no port-conflict failures.

**Single-shot server**: The server shuts down immediately after receiving one `POST /approve`. There is no way to approve multiple times or keep it running.

**Silent browser open**: `openBrowser` calls the platform-appropriate command (`open`, `start`, or `xdg-open`) and ignores errors. The URL printed to the terminal is always the reliable fallback.

**Atomic writes on approval**: Approved content is written via `tmp → os.Rename` to prevent partial writes if the process is killed mid-write — same pattern used throughout the project.

**Secondary artifact (PRD stage only)**: `tasks.yaml` is served alongside `PRD.md` when present. Both are written back on approval if the POST body includes non-empty `secondaryContent`.
