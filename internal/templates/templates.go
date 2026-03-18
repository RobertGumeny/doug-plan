package templates

import "embed"

// Init holds files copied into a project by `doug-plan init`.
//
//go:embed all:init
var Init embed.FS
