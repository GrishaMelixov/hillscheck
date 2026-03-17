package hillscheck

import "embed"

// StaticFiles holds the compiled React frontend (web/dist)
// and SQL migrations (migrations/).
// web/dist is populated at build time by the multi-stage Dockerfile.
//
//go:embed web/dist migrations
var StaticFiles embed.FS
