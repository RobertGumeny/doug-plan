package templates

import "embed"

// Init holds files copied into a project by `doug-plan init`.
//
//go:embed all:init
var Init embed.FS

// Steps holds ACTIVE_STEP.md templates for each pipeline stage.
// Files are named <Stage>.md (e.g. Discovery.md, Roadmapping.md).
//
//go:embed steps
var Steps embed.FS

// Artifacts holds host-owned artifact shell templates.
// Files are named after the artifact they seed (e.g. VISION.md, ROADMAP.md).
// The host writes these shells to disk before invoking the agent so the agent
// knows the required structure without needing to invent it.
//
//go:embed artifacts
var Artifacts embed.FS
