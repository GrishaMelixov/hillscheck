package hillscheck

// migrationsFS is embedded alongside StaticFiles so that main.go can access
// both web/dist and migrations via the same package-level embed.FS.
// The //go:embed directive in embed.go already covers web/dist;
// this file adds migrations/ to the same FS.
//
// NOTE: both files must be in the same package for the embed vars to merge.
// They are kept separate for clarity.
