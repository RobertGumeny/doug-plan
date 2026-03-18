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
