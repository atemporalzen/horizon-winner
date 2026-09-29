package web

import "embed"

// Files are embedded so runtime behavior never depends on the launch directory.
//
//go:embed index.html run.html popup.html controller.js runner.js engine.js style.css
var Files embed.FS
