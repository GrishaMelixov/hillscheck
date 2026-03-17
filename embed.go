package hillscheck

import "embed"

// StaticFiles holds the compiled React frontend (web/dist).
// It is populated at build time by the multi-stage Dockerfile
// (or locally by running `npm run build` inside the web/ directory).
//
//go:embed web/dist
var StaticFiles embed.FS
