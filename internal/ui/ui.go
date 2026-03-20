package ui

import "embed"

// Bundle holds the self-contained HTML bundle served by the embedded HTTP server.
// It is a placeholder that will be replaced by the compiled React bundle.
//
//go:embed bundle.html
var Bundle embed.FS
